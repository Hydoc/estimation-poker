package game

import (
	"github.com/google/uuid"
)

type Rooms interface {
	Save(r *Room)
	Find(id uuid.UUID) *Room
}

type RoomModel struct {
	rooms map[uuid.UUID]*Room
}

type Room struct {
	Id      uuid.UUID
	clients map[*Player]bool
}

func NewRoom(id uuid.UUID) *Room {
	return &Room{
		Id:      id,
		clients: make(map[*Player]bool),
	}
}

func (m *RoomModel) Save(r *Room) {
	m.rooms[r.Id] = r
}

func (m *RoomModel) Find(id uuid.UUID) *Room {
	if r, ok := m.rooms[id]; ok {
		return r
	}
	return nil
}

func NewRoomModel() *RoomModel {
	return &RoomModel{
		rooms: make(map[uuid.UUID]*Room),
	}
}
