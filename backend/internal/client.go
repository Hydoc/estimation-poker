package internal

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Hydoc/go-message"
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
	guess      int
	doSkip     bool
	send       chan *Message
	bus        message.Bus
}

func NewClient(name, role string, room *Room, connection *websocket.Conn, bus message.Bus) *Client {
	return &Client{
		room:       room,
		Name:       name,
		connection: connection,
		Role:       role,
		send:       make(chan *Message),
		bus:        bus,
	}
}

func HandleGuess(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(GuessPayload)
	if ok && payload.client.Role == Developer {
		payload.client.guess = payload.guess
		payload.client.doSkip = false
		payload.client.room.Broadcast <- newDeveloperGuessed()
		payload.client.send <- newYouGuessed(payload.guess)
	}
	return nil, nil
}

func HandleSkipRound(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(SkipRoundPayload)
	if ok && payload.client.Role == Developer {
		payload.client.doSkip = true
		payload.client.guess = 0
		payload.client.room.Broadcast <- newDeveloperSkipped()
		payload.client.send <- newYouSkipped()
	}
	return nil, nil
}

func HandleNewRound(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(NewRoundPayload)
	if ok && payload.client.Role == ProductOwner {
		payload.client.room.Broadcast <- newNewRound()
	}
	return nil, nil
}

func HandleEstimate(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(EstimatePayload)
	if ok && payload.client.Role == ProductOwner {
		payload.client.room.Broadcast <- newEstimate(payload.ticket)
	}
	return nil, nil
}

func HandleReveal(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(RevealPayload)
	if ok && payload.client.Role == ProductOwner {
		payload.client.room.Broadcast <- newReveal(payload.client.room.Clients)
	}
	return nil, nil
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

		cmd, err := fabricate(incMessage, client)
		if err != nil {
			log.Println(err)
			continue
		}
		client.bus.Dispatch(cmd)
		continue

		switch {
		// case incMessage.Type == SkipRound && client.Role == Developer:
		// 	client.doSkip = true
		// 	client.guess = 0
		// 	client.room.Broadcast <- newDeveloperSkipped()
		// 	client.send <- newYouSkipped()
		// case incMessage.Type == Guess && client.Role == Developer:
		// 	actualGuess := int(incMessage.Data.(float64))
		// 	client.guess = actualGuess
		// 	client.doSkip = false
		// 	client.room.Broadcast <- newDeveloperGuessed()
		// 	client.send <- newYouGuessed(actualGuess)
		// case incMessage.Type == NewRound && client.Role == ProductOwner:
		// 	client.room.Broadcast <- incMessage
		// case incMessage.Type == Reveal && client.Role == ProductOwner:
		// 	client.room.Broadcast <- newReveal(client.room.Clients)
		// case incMessage.Type == Estimate && client.Role == ProductOwner:
		// 	client.room.Broadcast <- incMessage
		case incMessage.Type == lockRoom:
			pw, pwOk := incMessage.Data.(map[string]any)["password"]
			key, keyOk := incMessage.Data.(map[string]any)["key"]

			if !keyOk {
				log.Printf("client: %s tried to lock room %s without a key\n", client.Name, client.room.Id)
				break
			}
			if !pwOk {
				log.Printf("client: %s tried to lock room %s without a password\n", client.Name, client.room.Id)
				break
			}

			if client.room.lock(client.Name, pw.(string), key.(string)) {
				client.room.Broadcast <- newRoomLocked()
				break
			}
			log.Println("was not able to lock room")
		case incMessage.Type == openRoom:
			key, keyOk := incMessage.Data.(map[string]any)["key"]

			if !keyOk {
				log.Println("client:", client.Name, "tried to open room", client.room.Id, "without a key")
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

func fabricate(incomingMessage *Message, client *Client) (message.Message, error) {
	switch incomingMessage.Type {
	case SkipRound:
		return message.New(
			SkipRound,
			SkipRoundPayload{
				client: client,
			},
		), nil
	case Estimate:
		ticket, ok := incomingMessage.Data.(string)
		if !ok {
			return message.Message{}, errors.New("ticket is invalid")
		}

		return message.New(
			Estimate,
			EstimatePayload{
				client: client,
				ticket: ticket,
			},
		), nil
	case Guess:
		actualGuess, ok := incomingMessage.Data.(float64)
		if !ok {
			return message.Message{}, errors.New("guess is invalid")
		}

		return message.New(Guess, GuessPayload{
			client: client,
			guess:  int(actualGuess),
		}), nil
	case NewRound:
		return message.New(NewRound, NewRoundPayload{client: client}), nil
	case Reveal:
		return message.New(Reveal, RevealPayload{client: client}), nil
	default:
		return message.Message{}, errors.New("command not found")
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

func (client *Client) newRound() {
	client.guess = 0
	client.doSkip = false
}

func (client *Client) asReveal() map[string]any {
	return map[string]any{
		"name":   client.Name,
		"role":   client.Role,
		"guess":  client.guess,
		"doSkip": client.doSkip,
	}
}

func (client *Client) ToJson() UserDTO {
	if client.Role == Developer {
		return map[string]any{
			"name":   client.Name,
			"role":   client.Role,
			"isDone": client.guess > 0 || client.doSkip,
		}
	}
	return map[string]any{
		"name": client.Name,
		"role": client.Role,
	}
}
