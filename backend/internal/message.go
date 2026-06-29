package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/Hydoc/go-message"
	"github.com/google/uuid"
)

const (
	leave           = "leave"
	guess           = "guess"
	newRound        = "new-round"
	estimate        = "estimate"
	lockRoom        = "lock-room"
	openRoom        = "open-room"
	skipRound       = "skip"
	reveal          = "reveal"
	roomLocked      = "room-locked"
	roomOpened      = "room-opened"
	developerAction = "developer-action"
	everyoneDone    = "everyone-done"
	youSkipped      = "you-skipped"
	youGuessed      = "you-guessed"
	addIssue        = "add-issue"
	issues          = "issues"
	permissions     = "permissions"
	users           = "users"
)

type IncomingWebsocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OutgoingWebsocketMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type SkipRoundPayload struct {
	client *Client
}

type NewRoundPayload struct {
	client *Client
}

type LockRoomPayload struct {
	client   *Client
	key      string
	password string
}

type OpenRoomPayload struct {
	client *Client
	key    string
}

type EstimatePayload struct {
	client *Client
	ticket string
}

type AddIssuePayload struct {
	client *Client
	issue  string
}

type GuessPayload struct {
	client *Client
	guess  int
}

type RevealPayload struct {
	client *Client
}

func newPermissions(clientName, roomCreatorName string, key uuid.UUID) *OutgoingWebsocketMessage {
	if clientName == roomCreatorName {
		return &OutgoingWebsocketMessage{
			Type: permissions,
			Data: Permissions{
				CanLockRoom: true,
				Key:         key.String(),
			},
		}
	}

	return &OutgoingWebsocketMessage{
		Type: permissions,
		Data: Permissions{
			CanLockRoom: false,
		},
	}
}

func newOutgoingWebsocketMessage(msgType string, data any) *OutgoingWebsocketMessage {
	return &OutgoingWebsocketMessage{
		Type: msgType,
		Data: data,
	}
}

func newUsers(clients map[*Client]bool) *OutgoingWebsocketMessage {
	var out []*Client

	for c := range clients {
		out = append(out, c)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return &OutgoingWebsocketMessage{
		Type: users,
		Data: out,
	}
}

func newReveal(clients map[*Client]bool) *OutgoingWebsocketMessage {
	out := []map[string]any{}
	for client := range clients {
		if client.Role == Developer {
			out = append(out, client.asReveal())
		}
	}

	return &OutgoingWebsocketMessage{
		Type: reveal,
		Data: out,
	}
}

func CreateBus() message.Bus {
	bus := message.NewBus()
	bus.Register(skipRound, handleSkipRound)
	bus.Register(estimate, handleEstimate)
	bus.Register(guess, handleGuess)
	bus.Register(newRound, handleNewRound)
	bus.Register(reveal, handleReveal)
	bus.Register(lockRoom, handleLockRoom)
	bus.Register(openRoom, handleOpenRoom)
	bus.Register(addIssue, handleAddIssue)
	return bus
}

func fabricate(incomingMessage *IncomingWebsocketMessage, client *Client) (message.Message, error) {
	switch incomingMessage.Type {
	case skipRound:
		return message.New(
			skipRound,
			SkipRoundPayload{
				client: client,
			},
		), nil
	case estimate:
		var ticket string
		if err := json.Unmarshal(incomingMessage.Data, &ticket); err != nil {
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
		var actualGuess int
		if err := json.Unmarshal(incomingMessage.Data, &actualGuess); err != nil {
			return message.Message{}, errors.New("guess is invalid")
		}
		return message.New(guess, GuessPayload{
			client: client,
			guess:  actualGuess,
		}), nil
	case newRound:
		return message.New(newRound, NewRoundPayload{client: client}), nil
	case reveal:
		return message.New(reveal, RevealPayload{client: client}), nil
	case lockRoom:
		var input struct {
			Password string `json:"password"`
			Key      string `json:"key"`
		}

		if err := json.Unmarshal(incomingMessage.Data, &input); err != nil {
			return message.Message{}, errors.New("lockRoom payload is invalid")
		}

		return message.New(lockRoom, LockRoomPayload{
			client:   client,
			key:      input.Key,
			password: input.Password,
		}), nil
	case openRoom:
		var input struct {
			Key string `json:"key"`
		}
		if err := json.Unmarshal(incomingMessage.Data, &input); err != nil {
			return message.Message{}, fmt.Errorf("client: %s tried to open room %s without a key", client.Name, client.room.Id)
		}

		return message.New(openRoom, OpenRoomPayload{
			client: client,
			key:    input.Key,
		}), nil
	case addIssue:
		var issue string
		if err := json.Unmarshal(incomingMessage.Data, &issue); err != nil {
			return message.Message{}, fmt.Errorf("client: %s tried to open room %s without a key", client.Name, client.room.Id)
		}

		return message.New(addIssue, AddIssuePayload{
			client: client,
			issue:  issue,
		}), nil
	default:
		return message.Message{}, errors.New("message not found")
	}
}
