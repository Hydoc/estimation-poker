package internal

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	ProductOwner = "product-owner"
	Developer    = "developer"
	pongWait     = 60 * time.Second
	pingPeriod   = (pongWait * 9) / 10
	writeWait    = 10 * time.Second
)

type userDTO map[string]any

type Client struct {
	connection *websocket.Conn
	room       *Room
	Name       string
	Role       string
	Guess      int
	DoSkip     bool
	send       chan message
}

func newClient(name, role string, room *Room, connection *websocket.Conn) *Client {
	return &Client{
		room:       room,
		Name:       name,
		connection: connection,
		Role:       role,
		send:       make(chan message),
	}
}

func (client *Client) websocketReader() {
	defer func() {
		client.room.leave <- client
		client.room.broadcast <- newLeave()
		client.connection.Close()
	}()
	client.connection.SetReadDeadline(time.Now().Add(pongWait))
	client.connection.SetPongHandler(func(string) error {
		client.connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var incMessage clientMessage
		err := client.connection.ReadJSON(&incMessage)
		if err != nil {
			log.Println("error reading incoming client message:", err)
			break
		}

		switch {
		case incMessage.Type == skipRound && client.Role == Developer:
			client.DoSkip = true
			client.room.broadcast <- newSkipRound()
			client.send <- newYouSkipped()
		case incMessage.Type == guess && client.Role == Developer:
			actualGuess := int(incMessage.Data.(float64))
			client.Guess = actualGuess
			client.room.broadcast <- newDeveloperGuessed()
			client.send <- newYouGuessed(actualGuess)
		case incMessage.Type == newRound && client.Role == ProductOwner:
			client.room.broadcast <- newResetRound()
		case incMessage.Type == lockRoom:
			pw, pwOk := incMessage.Data.(map[string]any)["password"]
			key, keyOk := incMessage.Data.(map[string]any)["key"]

			if !keyOk {
				log.Println(fmt.Sprintf("client: %s tried to lock room %s without a key", client.Name, client.room.id))
				break
			}
			if !pwOk {
				log.Println(fmt.Sprintf("client: %s tried to lock room %s without a password", client.Name, client.room.id))
				break
			}

			if client.room.lock(client.Name, pw.(string), key.(string)) {
				client.room.broadcast <- newRoomLocked()
				break
			}
			log.Println("was not able to lock room")
		case incMessage.Type == openRoom:
			key, keyOk := incMessage.Data.(map[string]any)["key"]

			if !keyOk {
				log.Println("client:", client.Name, "tried to open room", client.room.id, "without a key")
				break
			}

			if client.room.open(client.Name, key.(string)) {
				client.room.broadcast <- newRoomOpened()
				break
			}
			log.Println("was not able to open room")
		default:
			client.room.broadcast <- incMessage
		}
	}
}

func (client *Client) websocketWriter() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		client.connection.Close()
		ticker.Stop()
	}()
	for {
		select {
		case msg := <-client.send:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			err := client.connection.WriteJSON(msg.ToJson())
			if err != nil {
				log.Println("error writing to client:", err)
				return
			}
		case <-ticker.C:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) reset() {
	client.Guess = 0
	client.DoSkip = false
}

func (client *Client) toJson() userDTO {
	if client.Role == Developer {
		return map[string]interface{}{
			"name":   client.Name,
			"role":   client.Role,
			"guess":  client.Guess,
			"doSkip": client.DoSkip,
		}
	}
	return map[string]interface{}{
		"name": client.Name,
		"role": client.Role,
	}
}
