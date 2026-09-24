package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	roundJoin  = "round:join"
	roundLeave = "round:leave"

	issueAdd = "issue:add"

	issuesAll = "issues:all"
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

		if err := msg.Validate(); err != nil {
			return outgoingMessage{}, fmt.Errorf("validation failed: %w", err)
		}

		return handler(room, msg)
	}
}

type message interface {
	Validate() error
}

type issueAddMessage struct {
	Title string `json:"title"`
}

func (i issueAddMessage) Validate() error {
	i.Title = strings.TrimSpace(i.Title)

	if i.Title == "" {
		return errors.New("missing title")
	}

	if len(i.Title) > 15 {
		return errors.New("title too long")
	}

	return nil
}

func handleIssueAddMessage(room *room, msg issueAddMessage) (outgoingMessage, error) {
	room.addIssue(newIssue(msg.Title))

	return outgoingMessage{
		Type: issuesAll,
		Data: room.Issues(),
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
