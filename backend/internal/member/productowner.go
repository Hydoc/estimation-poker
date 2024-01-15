package member

import (
	"github.com/gorilla/websocket"
	"log"
)

type ProductOwner struct {
	clientInformation *ClientInformation
}

func (productOwner ProductOwner) Send(message []byte) {
	productOwner.clientInformation.connection.WriteMessage(websocket.TextMessage, message)
}

func (productOwner ProductOwner) WebsocketReader(broadcastInRoom func(roomId, message string)) {
	for {
		messageType, message, err := productOwner.clientInformation.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			broadcastInRoom(productOwner.clientInformation.RoomId, "leave")
			productOwner.clientInformation.connection.Close()
			break
		}
		if messageType == websocket.CloseMessage {
			broadcastInRoom(productOwner.clientInformation.RoomId, "leave")
			productOwner.clientInformation.connection.Close()
			break
		}
		log.Printf("receive: %s (type %d)", message, messageType)
		err = productOwner.clientInformation.connection.WriteMessage(messageType, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (productOwner ProductOwner) RoomId() string {
	return productOwner.clientInformation.RoomId
}
