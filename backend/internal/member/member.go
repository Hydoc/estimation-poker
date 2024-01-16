package member

import (
	"github.com/gorilla/websocket"
)

type Member interface {
	Send(message []byte)
	WebsocketReader(broadcastChannel chan interface{})
	RoomId() string
	Name() string
	ToJson() UserDTO
}

type Message[T any] interface {
	Payload() T
}

type Leave struct {
	member Member
}

type UserDTO map[string]interface{}

type ClientInformation struct {
	Name       string
	RoomId     string
	connection *websocket.Conn
}

func (leave Leave) Payload() Member {
	return leave.member
}

func NewLeave(member Member) *Leave {
	return &Leave{
		member: member,
	}
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
