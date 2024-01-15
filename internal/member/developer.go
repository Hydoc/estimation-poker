package member

import (
	"github.com/gorilla/websocket"
	"log"
)

type Developer struct {
	clientInformation *ClientInformation
	Guess             int
}

func (developer *Developer) Send(message []byte) {
	developer.clientInformation.connection.WriteMessage(websocket.TextMessage, message)
}

func (developer *Developer) Reader(broadcastInRoom func(roomId, message string)) {
	for {
		messageType, message, err := developer.clientInformation.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			broadcastInRoom(developer.clientInformation.RoomId, "leave")
			developer.clientInformation.connection.Close()
			break
		}
		if messageType == websocket.CloseMessage {
			broadcastInRoom(developer.clientInformation.RoomId, "leave")
			developer.clientInformation.connection.Close()
			break
		}
		log.Printf("receive: %s (type %d)", message, messageType)
		err = developer.clientInformation.connection.WriteMessage(messageType, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (developer *Developer) DoGuess(value int) {
	developer.Guess = value
}

func (developer *Developer) RoomId() string {
	return developer.clientInformation.RoomId
}
