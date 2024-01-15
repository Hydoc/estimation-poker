package member

import (
	"github.com/gorilla/websocket"
)

type Member interface {
	Send(message []byte)
	WebsocketReader(broadcastInRoom func(roomId, message string), removeFromRoom func(m Member))
	RoomId() string
	Name() string
	ToJson() UserDTO
}

type UserDTO map[string]interface{}

type ClientInformation struct {
	Name       string
	RoomId     string
	connection *websocket.Conn
}

func NewProductOwner(name, room string, connection *websocket.Conn) *ProductOwner {
	return &ProductOwner{
		&ClientInformation{
			RoomId:     room,
			Name:       name,
			connection: connection,
		},
	}
}

func NewDeveloper(name, room string, connection *websocket.Conn) *Developer {
	return &Developer{
		Guess: 0,
		clientInformation: &ClientInformation{
			RoomId:     room,
			Name:       name,
			connection: connection,
		},
	}
}
