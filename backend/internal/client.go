package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

type Permissions struct {
	CanLockRoom bool   `json:"canLockRoom"`
	Key         string `json:"key"`
}

type Client struct {
	connection *websocket.Conn
	logger     *slog.Logger
	room       *Room
	Name       string
	Role       string
	guess      int
	doSkip     bool
	send       chan *Message
	bus        message.Bus
}

func NewClient(name, role string, room *Room, connection *websocket.Conn, bus message.Bus, logger *slog.Logger) *Client {
	return &Client{
		room:       room,
		Name:       name,
		connection: connection,
		Role:       role,
		send:       make(chan *Message),
		bus:        bus,
		logger:     logger,
	}
}

func handleGuess(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(GuessPayload)
	if ok && payload.client.Role == Developer {
		payload.client.guess = payload.guess
		payload.client.doSkip = false
		payload.client.room.broadcast <- newDeveloperGuessed()
		payload.client.send <- newYouGuessed(payload.guess)
	}
	return nil, nil
}

func handleSkipRound(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(SkipRoundPayload)
	if ok && payload.client.Role == Developer {
		payload.client.doSkip = true
		payload.client.guess = 0
		payload.client.room.broadcast <- newDeveloperSkipped()
		payload.client.send <- newYouSkipped()
	}
	return nil, nil
}

func handleNewRound(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(NewRoundPayload)
	if ok && payload.client.Role == ProductOwner {
		payload.client.room.broadcast <- newNewRound()
	}
	return nil, nil
}

func handleLockRoom(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(LockRoomPayload)
	if ok && payload.client.room.lock(payload.client.Name, payload.password, payload.key) {
		payload.client.room.broadcast <- newRoomLocked()
	}
	return nil, nil
}

func handleOpenRoom(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(OpenRoomPayload)
	if ok && payload.client.room.open(payload.client.Name, payload.key) {
		payload.client.room.broadcast <- newRoomOpened()
	}
	return nil, nil
}

func handleEstimate(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(EstimatePayload)
	if ok && payload.client.Role == ProductOwner {
		payload.client.room.broadcast <- newEstimate(payload.ticket)
	}
	return nil, nil
}

func handleReveal(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(RevealPayload)
	if ok && payload.client.Role == ProductOwner {
		payload.client.room.broadcast <- newReveal(payload.client.room.Clients)
	}
	return nil, nil
}

func handleAddIssue(msg message.Message) (*message.Message, error) {
	payload, ok := msg.Payload.(AddIssuePayload)
	if ok && payload.client.Role == ProductOwner {
		payload.client.room.addIssue(payload.issue)
		payload.client.room.broadcast <- newIssues()
	}
	return nil, nil
}

func (client *Client) WebsocketReader() {
	defer func() {
		client.room.leave <- client
		client.room.broadcast <- newLeave(client.Name)
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
				client.logger.Error("error reading incoming client Message:", "error", err)
				return
			}
		}

		cmd, err := fabricate(incMessage, client)
		if err != nil {
			client.logger.Error(err.Error())
			continue
		}

		err = client.bus.Dispatch(cmd)
		if err != nil {
			client.logger.Error(err.Error())
		}
	}
}

func fabricate(incomingMessage *Message, client *Client) (message.Message, error) {
	switch incomingMessage.Type {
	case skipRound:
		return message.New(
			skipRound,
			SkipRoundPayload{
				client: client,
			},
		), nil
	case estimate:
		ticket, ok := incomingMessage.Data.(string)
		if !ok {
			return message.Message{}, errors.New("ticket is invalid")
		}

		return message.New(
			estimate,
			EstimatePayload{
				client: client,
				ticket: ticket,
			},
		), nil
	case guess:
		actualGuess, ok := incomingMessage.Data.(float64)
		if !ok {
			return message.Message{}, errors.New("guess is invalid")
		}

		return message.New(guess, GuessPayload{
			client: client,
			guess:  int(actualGuess),
		}), nil
	case newRound:
		return message.New(newRound, NewRoundPayload{client: client}), nil
	case reveal:
		return message.New(reveal, RevealPayload{client: client}), nil
	case lockRoom:
		pw, pwOk := incomingMessage.Data.(map[string]any)["password"]
		key, keyOk := incomingMessage.Data.(map[string]any)["key"]

		if !keyOk {
			return message.Message{}, fmt.Errorf("client: %s tried to lock room %s without a key", client.Name, client.room.Id)
		}
		if !pwOk {
			return message.Message{}, fmt.Errorf("client: %s tried to lock room %s without a password", client.Name, client.room.Id)
		}

		return message.New(lockRoom, LockRoomPayload{
			client:   client,
			key:      key.(string),
			password: pw.(string),
		}), nil
	case openRoom:
		key, keyOk := incomingMessage.Data.(map[string]any)["key"]

		if !keyOk {
			return message.Message{}, fmt.Errorf("client: %s tried to open room %s without a key", client.Name, client.room.Id)
		}

		return message.New(openRoom, OpenRoomPayload{
			client: client,
			key:    key.(string),
		}), nil
	case addIssue:
		actualIssue, ok := incomingMessage.Data.(string)
		if !ok {
			return message.Message{}, errors.New("issue is invalid")
		}

		return message.New(addIssue, AddIssuePayload{
			client: client,
			issue:  actualIssue,
		}), nil
	default:
		return message.Message{}, errors.New("message not found")
	}
}

func (client *Client) WebsocketWriter() {
	ticker := time.NewTicker(PingInterval)

	defer func() {
		client.room.leave <- client
		client.room.broadcast <- newLeave(client.Name)
		client.connection.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		select {
		case msg := <-client.send:
			err := wsjson.Write(context.Background(), client.connection, msg)
			if err != nil {
				client.logger.Error("error writing to client:", "error", err)
				return
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), PingInterval)
			err := client.connection.Ping(ctx)
			if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				cancel()
				client.logger.Error("error pinging client:", "error", err)
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
