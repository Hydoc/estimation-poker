package main

import "github.com/coder/websocket"

type subscriber struct {
	Name      string `json:"name"`
	messages  chan outgoingMessage
	closeSlow func()
}

func newSubscriber(name string, conn *websocket.Conn) *subscriber {
	return &subscriber{
		Name:     name,
		messages: make(chan outgoingMessage),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up")
		},
	}
}
