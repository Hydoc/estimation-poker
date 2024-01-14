package member

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

const (
	ProductOwner = "Product Owner"
	Developer    = "Developer"
)

type Member struct {
	Name       string
	Type       string
	RoomId     string
	Guess      int
	IsGuessing bool
	connection *websocket.Conn
}

func (member *Member) Send(message []byte) {
	member.connection.WriteMessage(websocket.TextMessage, message)
}

func (member *Member) Reader(broadcastInRoom func(roomId, message string)) {
	for {
		messageType, message, err := member.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			broadcastInRoom(member.RoomId, fmt.Sprintf("%s left.", member.Name))
			member.connection.Close()
			break
		}
		if messageType == websocket.CloseMessage {
			broadcastInRoom(member.RoomId, fmt.Sprintf("%s left.", member.Name))
			member.connection.Close()
			break
		}
		log.Printf("receive: %s (type %d)", message, messageType)
		err = member.connection.WriteMessage(messageType, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (member *Member) DoGuess(value int) {
	if member.Type != Developer {
		log.Fatal("member: member is not a developer")
		return
	}
	member.Guess = value
}

func NewMember(name, room, memberType string, connection *websocket.Conn) *Member {
	return &Member{
		RoomId:     room,
		Type:       memberType,
		Name:       name,
		IsGuessing: false,
		connection: connection,
	}
}
