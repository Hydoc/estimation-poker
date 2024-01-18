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
	err := developer.clientInformation.connection.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println("developer send:", err)
	}
}

func (developer *Developer) WebsocketReader(broadcastChannel chan Message) {
	for {
		messageType, incomingMessage, err := developer.clientInformation.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			developer.clientInformation.connection.Close()
			broadcastChannel <- NewLeave(developer)
			break
		}
		if messageType == websocket.CloseMessage {
			developer.clientInformation.connection.Close()
			broadcastChannel <- NewLeave(developer)
			break
		}
		log.Printf("receive: %s (type %d)", incomingMessage, messageType)
		err = developer.clientInformation.connection.WriteMessage(messageType, incomingMessage)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

//func (developer *Developer) DoGuess(value int) {
//	developer.Guess = value
//}

func (developer *Developer) RoomId() string {
	return developer.clientInformation.RoomId
}

func (developer *Developer) Name() string {
	return developer.clientInformation.Name
}

func (developer *Developer) ToJson() UserDTO {
	return map[string]interface{}{
		"name":  developer.clientInformation.Name,
		"role":  "developer",
		"guess": developer.Guess,
	}
}
