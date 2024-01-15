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

func (productOwner ProductOwner) WebsocketReader(broadcastInRoom func(roomId, message string), removeFromRoom func(m Member)) {
	for {
		messageType, message, err := productOwner.clientInformation.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			productOwner.clientInformation.connection.Close()
			removeFromRoom(productOwner)
			broadcastInRoom(productOwner.clientInformation.RoomId, "leave")
			break
		}
		if messageType == websocket.CloseMessage {
			productOwner.clientInformation.connection.Close()
			removeFromRoom(productOwner)
			broadcastInRoom(productOwner.clientInformation.RoomId, "leave")
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

func (productOwner ProductOwner) Name() string {
	return productOwner.clientInformation.Name
}

func (productOwner ProductOwner) ToJson() UserDTO {
	return map[string]interface{}{
		"name": productOwner.clientInformation.Name,
		"role": "product-owner",
	}
}
