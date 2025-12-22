package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RoomId string

type Room struct {
	clientMu sync.Mutex
	logger   *slog.Logger

	Id             RoomId
	InProgress     bool
	leave          chan *Client
	Join           chan *Client
	Clients        map[*Client]bool
	Broadcast      chan *Message
	destroy        chan<- RoomId
	NameOfCreator  string
	Key            uuid.UUID
	HashedPassword []byte
	Created        time.Time
}

func (room *Room) MarshalJSON() ([]byte, error) {
	out := struct {
		Id          RoomId `json:"id"`
		PlayerCount int    `json:"playerCount"`
	}{
		Id:          room.Id,
		PlayerCount: len(room.Clients),
	}
	return json.Marshal(&out)
}

func NewRoom(name RoomId, destroy chan<- RoomId, nameOfCreator string, logger *slog.Logger) *Room {
	return &Room{
		Id:             name,
		logger:         logger,
		InProgress:     false,
		leave:          make(chan *Client),
		Join:           make(chan *Client),
		Clients:        make(map[*Client]bool),
		Broadcast:      make(chan *Message),
		destroy:        destroy,
		NameOfCreator:  nameOfCreator,
		Key:            uuid.New(),
		HashedPassword: make([]byte, 0),
		Created:        time.Now(),
	}
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

func (room *Room) Verify(password string) bool {
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
		case client := <-room.Join:
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
		case msg := <-room.Broadcast:
			switch msg.Type {
			case Estimate:
				room.InProgress = true
				room.broadcastToClients(msg)
			case developerGuessed, SkipRound, developerSkipped:
				if room.everyDevIsDone() {
					room.broadcastToClients(newEveryoneIsDone())
					continue
				}
				room.broadcastToClients(msg)
			case NewRound:
				room.newRound()
			case leave:
				if room.InProgress {
					room.newRound()
					continue
				}
				room.broadcastToClients(msg)
			case join, Reveal, roomLocked, roomOpened:
				room.broadcastToClients(msg)
			default:
				room.logger.Error(fmt.Sprintf("unexpected Message %#v", msg))
			}
		}
	}
}
