package main

import (
	"encoding/json"
	"errors"
	"sync"
)

var (
	roundJoin  = "round:join"
	roundLeave = "round:leave"
)

type message interface {
	Validate() error

	ToBroadcastPayload() any
}

type roundJoinMessage struct {
}

func (j *roundJoinMessage) Validate() error {
	return nil
}

func (j *roundJoinMessage) ToBroadcastPayload() any {
	return map[string]any{
		"message": roundJoin,
	}
}

type roundLeaveMessage struct{}

func (r *roundLeaveMessage) Validate() error {
	return nil
}

func (r *roundLeaveMessage) ToBroadcastPayload() any {
	return map[string]any{
		"message": roundLeave,
	}
}

type factory func() message

var factories = map[string]factory{
	roundJoin:  func() message { return &roundJoinMessage{} },
	roundLeave: func() message { return &roundLeaveMessage{} },
}

type incomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func fabricateMessage(msg incomingMessage) (message, error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	foundFactory, exists := factories[msg.Type]
	if !exists {
		return nil, errors.New("unknown message")
	}

	cmd := foundFactory()

	if err := json.Unmarshal(msg.Data, cmd); err != nil {
		return nil, err
	}

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	return cmd, nil
}
