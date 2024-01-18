package member

import (
	"github.com/gorilla/websocket"
	"log"
)

type ProductOwner struct {
	clientInformation *ClientInformation
}

func (productOwner *ProductOwner) Send(message []byte) {
	err := productOwner.clientInformation.connection.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println("productowner send:", err)
	}
}

func (productOwner *ProductOwner) WebsocketReader(broadcastChannel chan Message) {
	for {
		messageType, incomingMessage, err := productOwner.clientInformation.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			productOwner.clientInformation.connection.Close()
			broadcastChannel <- NewLeave(productOwner)
			break
		}
		if messageType == websocket.CloseMessage {
			productOwner.clientInformation.connection.Close()
			broadcastChannel <- NewLeave(productOwner)
			break
		}
		log.Printf("receive: %s (type %d)", incomingMessage, messageType)
		err = productOwner.clientInformation.connection.WriteMessage(messageType, incomingMessage)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (productOwner *ProductOwner) RoomId() string {
	return productOwner.clientInformation.RoomId
}

func (productOwner *ProductOwner) Name() string {
	return productOwner.clientInformation.Name
}

func (productOwner *ProductOwner) ToJson() UserDTO {
	return map[string]interface{}{
		"name": productOwner.clientInformation.Name,
		"role": "product-owner",
	}
}
