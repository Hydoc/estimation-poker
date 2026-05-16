package internal

import (
	"errors"
	"fmt"

	"github.com/Hydoc/go-message"
	"github.com/google/uuid"
)

const (
	join             = "join"
	leave            = "leave"
	guess            = "guess"
	newRound         = "new-round"
	estimate         = "estimate"
	lockRoom         = "lock-room"
	openRoom         = "open-room"
	skipRound        = "skip"
	reveal           = "reveal"
	roomLocked       = "room-locked"
	roomOpened       = "room-opened"
	developerGuessed = "developer-guessed"
	everyoneDone     = "everyone-done"
	developerSkipped = "developer-skipped"
	youSkipped       = "you-skipped"
	youGuessed       = "you-guessed"
	addIssue         = "add-issue"
	issues           = "issues"
	permissions      = "permissions"
)

type WebsocketMessage struct {
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

func newPermissions(clientName, roomCreatorName string, key uuid.UUID) *WebsocketMessage {
	if clientName == roomCreatorName {
		return &WebsocketMessage{
			Type: permissions,
			Data: Permissions{
				CanLockRoom: true,
				Key:         key.String(),
			},
		}
	}

	return &WebsocketMessage{
		Type: permissions,
		Data: Permissions{
			CanLockRoom: false,
		},
	}
}

func newJoin() *WebsocketMessage {
	return &WebsocketMessage{
		Type: join,
	}
}

func newEstimate(ticket string) *WebsocketMessage {
	return &WebsocketMessage{
		Type: estimate,
		Data: ticket,
	}
}

func newLeave(name string) *WebsocketMessage {
	return &WebsocketMessage{
		Type: leave,
		Data: name,
	}
}

func newRoomLocked() *WebsocketMessage {
	return &WebsocketMessage{
		Type: roomLocked,
	}
}

func newRoomOpened() *WebsocketMessage {
	return &WebsocketMessage{
		Type: roomOpened,
	}
}

func newIssues() *WebsocketMessage {
	return &WebsocketMessage{
		Type: issues,
	}
}

func newDeveloperGuessed() *WebsocketMessage {
	return &WebsocketMessage{
		Type: developerGuessed,
	}
}

func newEveryoneIsDone() *WebsocketMessage {
	return &WebsocketMessage{
		Type: everyoneDone,
	}
}

func newReveal(clients map[*Client]bool) *WebsocketMessage {
	out := []map[string]any{}
	for client := range clients {
		if client.Role == Developer {
			out = append(out, client.asReveal())
		}
	}

	return &WebsocketMessage{
		Type: reveal,
		Data: out,
	}
}

func newNewRound() *WebsocketMessage {
	return &WebsocketMessage{
		Type: newRound,
	}
}

func newDeveloperSkipped() *WebsocketMessage {
	return &WebsocketMessage{
		Type: developerSkipped,
	}
}

func newYouSkipped() *WebsocketMessage {
	return &WebsocketMessage{
		Type: youSkipped,
	}
}

func newYouGuessed(guess int) *WebsocketMessage {
	return &WebsocketMessage{
		Type: youGuessed,
		Data: guess,
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

func fabricate(incomingMessage *WebsocketMessage, client *Client) (message.Message, error) {
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
