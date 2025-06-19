package game

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/Hydoc/guess-dev/backend/internal/message"
)

type Rooms interface {
	Save(r *Room)
	Find(id uuid.UUID) *Room
}

type RoomModel struct {
	rooms map[uuid.UUID]*Room
}

type Room struct {
	id      uuid.UUID
	clients map[*Player]bool
}

type CreateRoomMessage struct{}

func NewRoom() *Room {
	return &Room{
		id:      uuid.New(),
		clients: make(map[*Player]bool),
	}
}

func (m *RoomModel) Save(r *Room) {
	m.rooms[r.id] = r
}

func (m *RoomModel) Find(id uuid.UUID) *Room {
	if r, ok := m.rooms[id]; ok {
		return r
	}
	return nil
}

func CreateRoomHandler(messageChan <-chan message.Message) {
	for msg := range messageChan {
		actual, ok := msg.Payload.(CreateRoomMessage)
		if !ok {
			fmt.Println("invalid message data")
			continue
		}

		room := NewRoom()

	}
}

func NewRoomModel() *RoomModel {
	return &RoomModel{
		rooms: make(map[uuid.UUID]*Room),
	}
}
