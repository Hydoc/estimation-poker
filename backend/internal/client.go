package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Hydoc/go-message"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	ProductOwner = "product-owner"
	Developer    = "developer"
	PingInterval = time.Second * 20
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

func HandleLockRoom(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(LockRoomPayload)
	if ok && payload.client.room.lock(payload.client.Name, payload.password, payload.key) {
		payload.client.room.Broadcast <- newRoomLocked()
	}
	return nil, nil
}

func HandleOpenRoom(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(OpenRoomPayload)
	if ok && payload.client.room.open(payload.client.Name, payload.key) {
		payload.client.room.Broadcast <- newRoomOpened()
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
		client.room.Broadcast <- newLeave(client.Name)
		client.connection.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		var incMessage *Message
		err := wsjson.Read(context.Background(), client.connection, &incMessage)

		if err != nil {
			switch websocket.CloseStatus(err) {
			case websocket.StatusNoStatusRcvd, websocket.StatusGoingAway:
				return
			default:
				log.Println("error reading incoming client Message:", err)
				return
			}

		}

		cmd, err := fabricate(incMessage, client)
		if err != nil {
			log.Println(err)
			continue
		}
		client.bus.Dispatch(cmd)
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
	case LockRoom:
		pw, pwOk := incomingMessage.Data.(map[string]any)["password"]
		key, keyOk := incomingMessage.Data.(map[string]any)["key"]

		if !keyOk {
			return message.Message{}, fmt.Errorf("client: %s tried to lock room %s without a key", client.Name, client.room.Id)
		}
		if !pwOk {
			return message.Message{}, fmt.Errorf("client: %s tried to lock room %s without a password", client.Name, client.room.Id)
		}

		return message.New(LockRoom, LockRoomPayload{
			client:   client,
			key:      key.(string),
			password: pw.(string),
		}), nil
	case OpenRoom:
		key, keyOk := incomingMessage.Data.(map[string]any)["key"]

		if !keyOk {
			return message.Message{}, fmt.Errorf("client: %s tried to open room %s without a key", client.Name, client.room.Id)
		}

		return message.New(OpenRoom, OpenRoomPayload{
			client: client,
			key:    key.(string),
		}), nil
	default:
		return message.Message{}, errors.New("message not found")
	}
}

func (client *Client) WebsocketWriter() {
	ticker := time.NewTicker(PingInterval)

	defer func() {
		client.room.leave <- client
		client.room.Broadcast <- newLeave(client.Name)
		client.connection.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		select {
		case msg := <-client.send:
			err := wsjson.Write(context.Background(), client.connection, msg)
			if err != nil {
				log.Println("error writing to client:", err)
				return
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), PingInterval)
			err := client.connection.Ping(ctx)
			if err != nil {
				cancel()
				log.Println("error pinging client:", err)
				return
			}
			cancel()
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
