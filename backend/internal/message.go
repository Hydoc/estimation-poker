package internal

import (
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

type Message struct {
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

func newPermissions(clientName, roomCreatorName string, key uuid.UUID) *Message {
	if clientName == roomCreatorName {
		return &Message{
			Type: permissions,
			Data: Permissions{
				CanLockRoom: true,
				Key:         key.String(),
			},
		}
	}

	return &Message{
		Type: permissions,
		Data: Permissions{
			CanLockRoom: false,
		},
	}
}

func newJoin() *Message {
	return &Message{
		Type: join,
	}
}

func newEstimate(ticket string) *Message {
	return &Message{
		Type: estimate,
		Data: ticket,
	}
}

func newLeave(name string) *Message {
	return &Message{
		Type: leave,
		Data: name,
	}
}

func newRoomLocked() *Message {
	return &Message{
		Type: roomLocked,
	}
}

func newRoomOpened() *Message {
	return &Message{
		Type: roomOpened,
	}
}

func newIssues() *Message {
	return &Message{
		Type: issues,
	}
}

func newDeveloperGuessed() *Message {
	return &Message{
		Type: developerGuessed,
	}
}

func newEveryoneIsDone() *Message {
	return &Message{
		Type: everyoneDone,
	}
}

func newReveal(clients map[*Client]bool) *Message {
	out := []map[string]any{}
	for client := range clients {
		if client.Role == Developer {
			out = append(out, client.asReveal())
		}
	}

	return &Message{
		Type: reveal,
		Data: out,
	}
}

func newNewRound() *Message {
	return &Message{
		Type: newRound,
	}
}

func newDeveloperSkipped() *Message {
	return &Message{
		Type: developerSkipped,
	}
}

func newYouSkipped() *Message {
	return &Message{
		Type: youSkipped,
	}
}

func newYouGuessed(guess int) *Message {
	return &Message{
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
