package internal

import (
	"github.com/gorilla/websocket"
	"log"
)

const (
	ProductOwner = "product-owner"
	Developer    = "developer"
)

type userDTO map[string]interface{}

type Client struct {
	connection *websocket.Conn
	hub        *Hub
	Name       string
	RoomId     string
	Role       string
	Guess      int
	send       chan message
}

func newProductOwner(roomId, name string, hub *Hub, connection *websocket.Conn) *Client {
	return &Client{
		hub:        hub,
		RoomId:     roomId,
		Name:       name,
		connection: connection,
		Role:       ProductOwner,
		send:       make(chan message),
	}
}

func newDeveloper(roomId, name string, hub *Hub, connection *websocket.Conn) *Client {
	return &Client{
		hub:        hub,
		RoomId:     roomId,
		Name:       name,
		connection: connection,
		Role:       Developer,
		Guess:      0,
		send:       make(chan message),
	}
}

func (client *Client) websocketReader() {
	defer func() {
		client.hub.Unregister <- client
		client.hub.roomBroadcast <- newRoomBroadcast(client.RoomId, newLeave())
		client.connection.Close()
	}()
	for {
		var incMessage clientMessage
		err := client.connection.ReadJSON(&incMessage)
		if err != nil {
			log.Println("read:", err)
			break
		}

		if incMessage.Type == Guess && client.Role == Developer {
			actualGuess := int(incMessage.Data.(float64))
			client.Guess = actualGuess
			client.hub.roomBroadcast <- newRoomBroadcast(client.RoomId, newDeveloperGuessed())
			client.send <- newYouGuessed(actualGuess)
		} else if incMessage.Type == NewRound && client.Role == ProductOwner {
			client.hub.roomBroadcast <- newRoomBroadcast(client.RoomId, newResetRound())
		} else {
			client.hub.roomBroadcast <- newRoomBroadcast(client.RoomId, incMessage)
		}
	}
}

func (client *Client) websocketWriter() {
	defer func() {
		client.connection.Close()
	}()
	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			err := client.connection.WriteJSON(msg.ToJson())
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (client *Client) reset() {
	client.Guess = 0
}

func (client *Client) toJson() userDTO {
	if client.Role == Developer {
		return map[string]interface{}{
			"name":  client.Name,
			"role":  client.Role,
			"guess": client.Guess,
		}
	}
	return map[string]interface{}{
		"name": client.Name,
		"role": client.Role,
	}
}
