package main

import (
	"uuid"

	"github.com/coder/websocket"
)

type subscriber struct {
	Id        uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	messages  chan outgoingMessage
	closeSlow func()
}

func newSubscriber(name string, conn *websocket.Conn) *subscriber {
	return &subscriber{
		Id:       uuid.New(),
		Name:     name,
		messages: make(chan outgoingMessage),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up")
		},
	}
}
