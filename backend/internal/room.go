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

	id             RoomId
	inProgress     bool
	leave          chan *Client
	join           chan *Client
	clients        map[*Client]bool
	broadcast      chan message
	destroy        chan<- RoomId
	nameOfCreator  string
	isLocked       bool
	key            uuid.UUID
	hashedPassword []byte
}

func newRoom(name RoomId, destroy chan<- RoomId, nameOfCreator string) *Room {
	return &Room{
		id:             name,
		inProgress:     false,
		leave:          make(chan *Client),
		join:           make(chan *Client),
		clients:        make(map[*Client]bool),
		broadcast:      make(chan message),
		destroy:        destroy,
		nameOfCreator:  nameOfCreator,
		isLocked:       false,
		key:            uuid.New(),
		hashedPassword: make([]byte, 0),
	}
}

func (room *Room) lock(username, password, key string) bool {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("could not hash password %s\n", password)
		return false
	}
	if username == room.nameOfCreator && key == room.key.String() {
		room.isLocked = true
		room.hashedPassword = hashed
		return true
	}

	return false
}

func (room *Room) open(username, key string) bool {
	if username == room.nameOfCreator && key == room.key.String() {
		room.isLocked = false
		room.hashedPassword = make([]byte, 0)
		return true
	}
	return false
}

func (room *Room) verify(password string) bool {
	err := bcrypt.CompareHashAndPassword(room.hashedPassword, []byte(password))
	return err == nil
}

func (room *Room) everyDevIsDone() bool {
	for client := range room.clients {
		if client.Role == Developer && (client.Guess == 0 && !client.DoSkip) {
			return false
		}
	}
	return true
}

func (room *Room) resetRound(client *Client) {
	room.inProgress = false
	if client.Role == Developer {
		client.reset()
	}
	client.send <- newResetRound()
}

func (room *Room) Run() {
	for {
		select {
		case client := <-room.join:
			room.clientMu.Lock()
			room.clients[client] = true
			room.clientMu.Unlock()
		case client := <-room.leave:
			if _, ok := room.clients[client]; ok {
				delete(room.clients, client)
			}
			if len(room.clients) == 0 {
				room.destroy <- room.id
			}
		case msg := <-room.broadcast:
			for client := range room.clients {
				switch msg.Type {
				case estimate:
					room.inProgress = true
					client.send <- msg
				case developerGuessed, skipRound:
					if room.everyDevIsDone() {
						client.send <- newEveryoneIsDone()
						continue
					}
					client.send <- msg
				case resetRound:
					room.resetRound(client)
				case leave:
					if room.inProgress {
						for c := range room.clients {
							room.resetRound(c)
						}
						continue
					}
					client.send <- msg
				case join, revealRound:
					client.send <- msg
				default:
					log.Printf("unexpected message %#v", msg)
				}
			}
		}
	}
}
