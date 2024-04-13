package internal

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type RoomId string

type Room struct {
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
		log.Fatalf("could not hash password %s", password)
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
		if client.Role == Developer && (client.Guess == 0 || !client.DoSkip) {
			return false
		}
	}
	return true
}

func (room *Room) Run() {
	for {
		select {
		case client := <-room.join:
			room.clients[client] = true
		case client := <-room.leave:
			if _, ok := room.clients[client]; ok {
				delete(room.clients, client)
			}
			if len(room.clients) == 0 {
				room.destroy <- room.id
			}
		case msg := <-room.broadcast:
			for client := range room.clients {
				switch msg.(type) {
				case clientMessage:
					if msg.(clientMessage).isEstimate() {
						room.inProgress = true
					}
					client.send <- msg
				case developerGuessed, skip:
					if room.everyDevIsDone() {
						client.send <- newEveryoneIsDone()
						continue
					}
					client.send <- msg
				case resetRound:
					room.inProgress = false
					if client.Role == Developer {
						client.reset()
					}
					client.send <- msg
				case leave:
					if room.inProgress {
						room.inProgress = false
						if client.Role == Developer {
							client.reset()
						}
						client.send <- newResetRound()
						continue
					}
					client.send <- msg
				default:
					client.send <- msg
				}
			}
		}
	}
}
