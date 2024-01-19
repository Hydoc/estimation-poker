package member

import (
	"github.com/gorilla/websocket"
)

type Member interface {
	Send(message []byte)
	WebsocketReader(broadcastChannel chan Message)
	RoomId() string
	Name() string
	ToJson() UserDTO
}

type Message interface {
	ToJson() MessageDTO
}

type IncomingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (incMessage IncomingMessage) ToJson() MessageDTO {
	return map[string]interface{}{
		"type": incMessage.Type,
		"data": incMessage.Data,
	}
}

type Leave struct {
	member Member
}

type Join struct{}

type DeveloperGuessed struct {
	Guess int
}

type EveryoneGuessed struct{}

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

func (devGuessed DeveloperGuessed) ToJson() MessageDTO {
	return map[string]interface{}{
		"type": "developer-guessed",
		"data": devGuessed.Guess,
	}
}

func (everyoneGuessed EveryoneGuessed) ToJson() MessageDTO {
	return map[string]interface{}{
		"type": "everyone-guessed",
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

func NewDeveloperGuessed(guess int) DeveloperGuessed {
	return DeveloperGuessed{
		Guess: guess,
	}
}

func NewEveryoneGuessed() EveryoneGuessed {
	return EveryoneGuessed{}
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
