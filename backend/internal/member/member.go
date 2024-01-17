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

type Message interface {
	ToJson() MessageDTO
}

type Leave struct {
	member Member
}

type Join struct{}

type MessageDTO map[string]interface{}

type UserDTO map[string]interface{}

type ClientInformation struct {
	Name       string
	RoomId     string
	connection *websocket.Conn
}

func (join Join) ToJson() MessageDTO {
	return map[string]interface{}{
		"type": "join",
	}
}

func (leave Leave) Payload() Member {
	return leave.member
}

func (leave Leave) ToJson() MessageDTO {
	return map[string]interface{}{
		"type": "leave",
	}
}

func NewJoin() Join {
	return Join{}
}

func NewLeave(member Member) Leave {
	return Leave{
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
