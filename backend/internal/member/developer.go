package member

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

const (
	guess = "guess"
)

type Developer struct {
	clientInformation *clientInformation
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
		var developerMessage IncomingMessage
		err = json.Unmarshal(incomingMessage, &developerMessage)
		if err != nil {
			log.Printf("developer receive: could not unmarshal message %s", incomingMessage)
		}

		if developerMessage.Type == guess {
			actualGuess := int(developerMessage.Data.(float64))
			developer.Guess = actualGuess
			encoded, _ := json.Marshal(NewYouGuessed(actualGuess).ToJson())
			developer.Send(encoded)
			broadcastChannel <- NewDeveloperGuessed()
		}
	}
}

func (developer *Developer) Reset() {
	developer.Guess = 0
}

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
