package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Hydoc/estimation-poker/backend/internal/validator"
)

var (
	// declare incoming messages here
	roomJoin  = "room:join"
	roomLeave = "room:leave"
	issueAdd  = "issue:add"

	// declare outgoing messages here
	issuesAll = "issues:all"
	roomState = "room:state"
)

type HandlerFunc func(room *room, rawPayload json.RawMessage) (message outgoingMessage, err error)

type messageHandlerRegistry struct {
	handlersMu sync.RWMutex
	handlers   map[string]HandlerFunc
}

func newMessageHandlerRegistry() *messageHandlerRegistry {
	return &messageHandlerRegistry{
		handlers: make(map[string]HandlerFunc),
	}
}

func (r *messageHandlerRegistry) register[T message](messageType string, handler func(room *room, msg T) (message outgoingMessage, e error)) {
	r.handlers[messageType] = func(room *room, rawPayload json.RawMessage) (message outgoingMessage, e error) {
		var msg T
		if err := json.Unmarshal(rawPayload, &msg); err != nil {
			return outgoingMessage{}, fmt.Errorf("invalid payload: %w", err)
		}

		v := validator.New()

		if msg.Validate(v); !v.Valid() {
			return outgoingMessage{}, fmt.Errorf("validation failed")
		}

		return handler(room, msg)
	}
}

type message interface {
	Validate(v *validator.Validator)
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

func handleIssueAddMessage(room *room, msg issueAddMessage) (outgoingMessage, error) {
	room.addIssue(newIssue(msg.Title))

	return outgoingMessage{
		Type: issuesAll,
		Data: room.Issues(),
	}, nil
}

func handleRoundJoinMessage(room *room, _ roundJoinMessage) (outgoingMessage, error) {
	return outgoingMessage{
		Type: roomState,
		Data: room.State(),
	}, nil
}

func handleRoundLeaveMessage(room *room, _ roundLeaveMessage) (outgoingMessage, error) {
	return outgoingMessage{
		Type: roomState,
		Data: room.State(),
	}, nil
}

type incomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type outgoingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
