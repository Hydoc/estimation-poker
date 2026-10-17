package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"uuid"

	"github.com/Hydoc/estimation-poker/backend/internal/validator"
)

var (
	// declare incoming messages here
	roomJoin      = "room:join"
	roomLeave     = "room:leave"
	issueAdd      = "issue:add"
	issueEstimate = "issue:estimate"

	// declare outgoing messages here
	issuesAll = "issues:all"
	roomState = "room:state"
)

type HandlerFunc func(room *room, inc incomingMessage) (message outgoingMessage, err error)

type messageHandlerRegistry struct {
	handlersMu sync.RWMutex
	handlers   map[string]HandlerFunc
}

func newMessageHandlerRegistry() *messageHandlerRegistry {
	return &messageHandlerRegistry{
		handlers: make(map[string]HandlerFunc),
	}
}

func (r *messageHandlerRegistry) register[T message](messageType string, handler func(room *room, userId uuid.UUID, msg T) (message outgoingMessage, e error)) {
	r.handlers[messageType] = func(room *room, inc incomingMessage) (message outgoingMessage, e error) {
		var msg T
		if err := json.Unmarshal(inc.Data, &msg); err != nil {
			return outgoingMessage{}, fmt.Errorf("invalid payload: %w", err)
		}

		v := validator.New()

		if msg.Validate(v); !v.Valid() {
			return outgoingMessage{}, fmt.Errorf("validation failed")
		}

		return handler(room, inc.UserId, msg)
	}
}

type message interface {
	Validate(v *validator.Validator)
}

type issueEstimateMessage struct {
	Id uuid.UUID `json:"id"`
}

func (e issueEstimateMessage) Validate(v *validator.Validator) {
	v.Check(uuid.Nil() != e.Id, "id", "must not be nil uuid")
}

type roundJoinMessage struct{}

func (r roundJoinMessage) Validate(v *validator.Validator) {}

type roundLeaveMessage struct{}

func (r roundLeaveMessage) Validate(v *validator.Validator) {}

type issueAddMessage struct {
	Title string `json:"title"`
}

func (i issueAddMessage) Validate(v *validator.Validator) {
	v.Check(i.Title != "", "title", "must be set")
	v.Check(len(i.Title) <= 15, "title", "must not be more than 15 bytes long")
}

func handleIssueEstimateMessage(room *room, _ uuid.UUID, msg issueEstimateMessage) (outgoingMessage, error) {
	room.mu.Lock()
	defer room.mu.Unlock()

	for _, iss := range room.issues {
		if iss.Id == msg.Id {
			room.activeIssue = iss
			room.SetInProgress(true)
			return outgoingMessage{Type: roomState, Data: room.State()}, nil
		}
	}

	return outgoingMessage{}, errors.New("issue not found")
}

func handleIssueAddMessage(room *room, _ uuid.UUID, msg issueAddMessage) (outgoingMessage, error) {
	room.addIssue(newIssue(msg.Title))

	return outgoingMessage{
		Type: issuesAll,
		Data: room.Issues(),
	}, nil
}

func handleRoundJoinMessage(room *room, _ uuid.UUID, _ roundJoinMessage) (outgoingMessage, error) {
	return outgoingMessage{
		Type: roomState,
		Data: room.State(),
	}, nil
}

func handleRoundLeaveMessage(room *room, _ uuid.UUID, _ roundLeaveMessage) (outgoingMessage, error) {
	return outgoingMessage{
		Type: roomState,
		Data: room.State(),
	}, nil
}

type incomingMessage struct {
	Type   string          `json:"type"`
	UserId uuid.UUID       `json:"userId"`
	Data   json.RawMessage `json:"data"`
}

type outgoingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
