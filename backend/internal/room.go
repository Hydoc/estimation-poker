package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken = errors.New("username already taken")
	ErrRoundStarted  = errors.New("round already started")
	ErrWrongPassword = errors.New("wrong password")
)

type RoomId string

type Issue struct {
	Title string
	Guess int
}

type Room struct {
	clientMu sync.Mutex
	logger   *slog.Logger

	Id             RoomId
	InProgress     bool
	leave          chan *Client
	join           chan *Client
	Clients        map[*Client]bool
	broadcast      chan *Message
	destroy        chan<- RoomId
	NameOfCreator  string
	Key            uuid.UUID
	HashedPassword []byte
	Created        time.Time
	Issues         []Issue
	GuessConfig    *GuessConfig
}

type ConnectionState struct {
	CanConnect bool   `json:"canConnect"`
	Reason     string `json:"reason"`
}

type State struct {
	InProgress      bool               `json:"inProgress"`
	IsLocked        bool               `json:"isLocked"`
	Issues          []Issue            `json:"issues"`
	PossibleGuesses []GuessConfigEntry `json:"possibleGuesses"`
}

type Overview struct {
	Id          RoomId    `json:"id"`
	PlayerCount int       `json:"playerCount"`
	Created     time.Time `json:"-"`
}

func (room *Room) State() State {
	return State{
		InProgress:      room.InProgress,
		IsLocked:        room.IsLocked(),
		Issues:          room.Issues,
		PossibleGuesses: room.GuessConfig.Guesses,
	}
}

func (room *Room) AsOverview() Overview {
	return Overview{
		Id:          room.Id,
		PlayerCount: len(room.Clients),
		Created:     room.Created,
	}
}

func NewRoom(name RoomId, destroy chan<- RoomId, nameOfCreator string, logger *slog.Logger, guessConfig *GuessConfig) *Room {
	return &Room{
		Id:             name,
		logger:         logger,
		InProgress:     false,
		leave:          make(chan *Client),
		join:           make(chan *Client),
		Clients:        make(map[*Client]bool),
		broadcast:      make(chan *Message),
		destroy:        destroy,
		NameOfCreator:  nameOfCreator,
		Key:            uuid.New(),
		HashedPassword: make([]byte, 0),
		Created:        time.Now(),
		Issues:         make([]Issue, 0),
		GuessConfig:    guessConfig,
	}
}

func (room *Room) Join(client *Client) {
	room.join <- client
	client.send <- newPermissions(client.Name, room.NameOfCreator, room.Key)
	room.broadcast <- newJoin()
}

func (room *Room) lock(username, password, key string) bool {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		room.logger.Error("failed to hash password")
		return false
	}
	if username == room.NameOfCreator && key == room.Key.String() {
		room.HashedPassword = hashed
		return true
	}

	return false
}

func (room *Room) open(username, key string) bool {
	if username == room.NameOfCreator && key == room.Key.String() {
		room.HashedPassword = make([]byte, 0)
		return true
	}
	return false
}

func (room *Room) ConnectionState(username string, password string) ConnectionState {
	if room.InProgress {
		return ConnectionState{
			CanConnect: false,
			Reason:     ErrRoundStarted.Error(),
		}
	}

	if room.IsLocked() && !room.verify(password) {
		return ConnectionState{
			CanConnect: false,
			Reason:     ErrWrongPassword.Error(),
		}
	}

	room.clientMu.Lock()
	defer room.clientMu.Unlock()

	for client := range room.Clients {
		if client.Name == username {
			return ConnectionState{
				CanConnect: false,
				Reason:     ErrUsernameTaken.Error(),
			}
		}
	}

	return ConnectionState{
		CanConnect: true,
		Reason:     "",
	}
}

func (room *Room) verify(password string) bool {
	err := bcrypt.CompareHashAndPassword(room.HashedPassword, []byte(password))
	return err == nil
}

func (room *Room) everyDevIsDone() bool {
	for client := range room.Clients {
		if client.Role == Developer && (client.guess == 0 && !client.doSkip) {
			return false
		}
	}
	return true
}

func (room *Room) newRound() {
	room.InProgress = false
	for client := range room.Clients {
		client.newRound()
		client.send <- newNewRound()
	}
}

func (room *Room) broadcastToClients(msg *Message) {
	for client := range room.Clients {
		client.send <- msg
	}
}

func (room *Room) IsLocked() bool {
	return len(room.HashedPassword) > 0
}

func (room *Room) Run() {
	for {
		select {
		case client := <-room.join:
			room.clientMu.Lock()
			room.Clients[client] = true
			room.clientMu.Unlock()
		case client := <-room.leave:
			room.clientMu.Lock()
			delete(room.Clients, client)
			if len(room.Clients) == 0 {
				room.destroy <- room.Id
			}
			room.clientMu.Unlock()
		case msg := <-room.broadcast:
			switch msg.Type {
			case estimate:
				room.InProgress = true
				room.broadcastToClients(msg)
			case developerGuessed, skipRound, developerSkipped:
				if room.everyDevIsDone() {
					room.broadcastToClients(newEveryoneIsDone())
					continue
				}
				room.broadcastToClients(msg)
			case newRound:
				room.newRound()
			case leave:
				if room.InProgress {
					room.newRound()
					continue
				}
				room.broadcastToClients(msg)
			case join, reveal, roomLocked, roomOpened:
				room.broadcastToClients(msg)
			default:
				room.logger.Error(fmt.Sprintf("unexpected Message %#v", msg))
			}
		}
	}
}

func (room *Room) addIssue(issue string) {
	room.Issues = append(room.Issues, Issue{
		Title: issue,
		Guess: -1,
	})
}
