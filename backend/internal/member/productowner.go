package member

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type ProductOwner struct {
	clientInformation *clientInformation
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
			broadcastChannel <- NewLeave(productOwner)
			productOwner.clientInformation.connection.Close()
			break
		}
		if messageType == websocket.CloseMessage {
			broadcastChannel <- NewLeave(productOwner)
			productOwner.clientInformation.connection.Close()
			break
		}
		var productOwnerMessage IncomingMessage
		err = json.Unmarshal(incomingMessage, &productOwnerMessage)
		if err != nil {
			log.Printf("productowner receive: could not unmarshal message %s", incomingMessage)
		}

		log.Println(productOwnerMessage, string(incomingMessage))
		broadcastChannel <- productOwnerMessage
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
