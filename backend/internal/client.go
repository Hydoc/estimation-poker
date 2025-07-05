package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	ProductOwner = "product-owner"
	Developer    = "developer"
	pongWait     = 60 * time.Second
	pingPeriod   = (pongWait * 9) / 10
	writeWait    = 10 * time.Second
)

type UserDTO map[string]any

type Client struct {
	connection *websocket.Conn
	room       *Room
	Name       string
	Role       string
	Guess      int
	DoSkip     bool
	send       chan *Message
}

func NewClient(name, role string, room *Room, connection *websocket.Conn) *Client {
	return &Client{
		room:       room,
		Name:       name,
		connection: connection,
		Role:       role,
		send:       make(chan *Message),
	}
}

func (client *Client) WebsocketReader() {
	defer func() {
		client.room.leave <- client
		client.room.Broadcast <- newLeave()
		client.connection.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		var incMessage *Message
		err := wsjson.Read(context.Background(), client.connection, &incMessage)
		if err != nil {
			log.Println("error reading incoming client Message:", err)
			return
		}

		switch {
		case incMessage.Type == skipRound && client.Role == Developer:
			client.DoSkip = true
			client.Guess = 0
			client.room.Broadcast <- newDeveloperSkipped()
			client.send <- newYouSkipped()
		case incMessage.Type == guess && client.Role == Developer:
			actualGuess := int(incMessage.Data.(float64))
			client.Guess = actualGuess
			client.DoSkip = false
			client.room.Broadcast <- newDeveloperGuessed()
			client.send <- newYouGuessed(actualGuess)
		case incMessage.Type == newRound && client.Role == ProductOwner:
			client.room.Broadcast <- incMessage
		case incMessage.Type == reveal && client.Role == ProductOwner:
			client.room.Broadcast <- newReveal(client.room.Clients)
		case incMessage.Type == estimate && client.Role == ProductOwner:
			client.room.Broadcast <- incMessage
		case incMessage.Type == lockRoom:
			pw, pwOk := incMessage.Data.(map[string]any)["password"]
			key, keyOk := incMessage.Data.(map[string]any)["Key"]

			if !keyOk {
				log.Println(fmt.Sprintf("client: %s tried to lock room %s without a Key", client.Name, client.room.Id))
				break
			}
			if !pwOk {
				log.Println(fmt.Sprintf("client: %s tried to lock room %s without a password", client.Name, client.room.Id))
				break
			}

			if client.room.lock(client.Name, pw.(string), key.(string)) {
				client.room.Broadcast <- newRoomLocked()
				break
			}
			log.Println("was not able to lock room")
		case incMessage.Type == openRoom:
			key, keyOk := incMessage.Data.(map[string]any)["Key"]

			if !keyOk {
				log.Println("client:", client.Name, "tried to open room", client.room.Id, "without a Key")
				break
			}

			if client.room.open(client.Name, key.(string)) {
				client.room.Broadcast <- newRoomOpened()
				break
			}
			log.Println("was not able to open room")
		default:
			log.Printf("unknown Message %#v\n", incMessage)
		}
	}
}

func (client *Client) WebsocketWriter() {
	for {
		select {
		case msg := <-client.send:
			err := wsjson.Write(context.Background(), client.connection, msg)
			if err != nil {
				log.Println("error writing to client:", err)
				return
			}
		}
	}
}

func (client *Client) reset() {
	client.Guess = 0
	client.DoSkip = false
}

func (client *Client) asReveal() map[string]any {
	return map[string]any{
		"name":   client.Name,
		"role":   client.Role,
		"guess":  client.Guess,
		"doSkip": client.DoSkip,
	}
}

func (client *Client) ToJson() UserDTO {
	if client.Role == Developer {
		return map[string]any{
			"name":   client.Name,
			"role":   client.Role,
			"isDone": client.Guess > 0 || client.DoSkip,
		}
	}
	return map[string]any{
		"name": client.Name,
		"role": client.Role,
	}
}
