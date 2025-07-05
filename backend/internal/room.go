package internal

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RoomId string

type Room struct {
	clientMu sync.Mutex

	Id             RoomId
	InProgress     bool
	leave          chan *Client
	Join           chan *Client
	Clients        map[*Client]bool
	Broadcast      chan *Message
	destroy        chan<- RoomId
	NameOfCreator  string
	IsLocked       bool
	Key            uuid.UUID
	HashedPassword []byte
}

func NewRoom(name RoomId, destroy chan<- RoomId, nameOfCreator string) *Room {
	return &Room{
		Id:             name,
		InProgress:     false,
		leave:          make(chan *Client),
		Join:           make(chan *Client),
		Clients:        make(map[*Client]bool),
		Broadcast:      make(chan *Message),
		destroy:        destroy,
		NameOfCreator:  nameOfCreator,
		IsLocked:       false,
		Key:            uuid.New(),
		HashedPassword: make([]byte, 0),
	}
}

func (room *Room) lock(username, password, key string) bool {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("could not hash password %s\n", password)
		return false
	}
	if username == room.NameOfCreator && key == room.Key.String() {
		room.IsLocked = true
		room.HashedPassword = hashed
		return true
	}

	return false
}

func (room *Room) open(username, key string) bool {
	if username == room.NameOfCreator && key == room.Key.String() {
		room.IsLocked = false
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
		if client.Role == Developer && (client.Guess == 0 && !client.DoSkip) {
			return false
		}
	}
	return true
}

func (room *Room) newRound(client *Client) {
	room.InProgress = false
	if client.Role == Developer {
		client.reset()
	}
	client.send <- newNewRound()
}

func (room *Room) Run() {
	for {
		select {
		case client := <-room.Join:
			room.clientMu.Lock()
			room.Clients[client] = true
			room.clientMu.Unlock()
		case client := <-room.leave:
			if _, ok := room.Clients[client]; ok {
				delete(room.Clients, client)
			}
			if len(room.Clients) == 0 {
				room.destroy <- room.Id
			}
		case msg := <-room.Broadcast:
			for client := range room.Clients {
				switch msg.Type {
				case estimate:
					room.InProgress = true
					client.send <- msg
				case developerGuessed, skipRound, developerSkipped:
					if room.everyDevIsDone() {
						client.send <- newEveryoneIsDone()
						continue
					}
					client.send <- msg
				case newRound:
					room.newRound(client)
				case leave:
					if room.InProgress {
						for c := range room.Clients {
							room.newRound(c)
						}
						continue
					}
					client.send <- msg
				case join, reveal:
					client.send <- msg
				default:
					log.Printf("unexpected Message %#v", msg)
				}
			}
		}
	}
}
