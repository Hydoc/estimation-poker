package internal

import (
	"reflect"
	"testing"
)

func TestHub_IsRoundInRoomInProgress_WhenRoundIsInProgress(t *testing.T) {
	roomId := "1"
	testRoomBroadCastChannel := make(chan roomBroadcastMessage)
	doneChannel := make(chan message)

	hub := Hub{
		roomBroadcast: testRoomBroadCastChannel,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
	go hub.Run()

	client := &Client{RoomId: roomId, send: doneChannel}
	hub.register <- client
	hub.roomBroadcast <- newRoomBroadcast(roomId, clientMessage{
		"estimate",
		"WR-123",
	})

	close(testRoomBroadCastChannel)

	// wait for send from hub to client
	<-doneChannel

	want := true
	got := hub.IsRoundInRoomInProgress(roomId)
	if got != want {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestHub_IsRoundInRoomInProgress_WhenRoundIsNotInProgress(t *testing.T) {
	hub := NewHub()

	want := false
	got := hub.IsRoundInRoomInProgress("1")
	if got != want {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestHub_Run_WhenRegisteringAClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	client := &Client{}
	hub.register <- client

	_, ok := hub.clients[client]
	if !ok {
		t.Error("expected to find client")
	}
}

func TestHub_Run_WhenUnregisteringAClient(t *testing.T) {
	client := &Client{}
	clients := make(map[*Client]bool)
	clients[client] = true

	hub := Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       clients,
		rooms:         make(map[string]bool),
	}
	go hub.Run()
	hub.unregister <- client

	_, ok := hub.clients[client]
	if ok {
		t.Error("expected to not find client")
	}
}

func TestHub_Run_WhenRoomBroadcastAClientMessage(t *testing.T) {
	clientChannel := make(chan message)
	client := &Client{RoomId: "1", send: clientChannel}
	clients := make(map[*Client]bool)
	clients[client] = true

	hub := Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       clients,
		rooms:         make(map[string]bool),
	}
	go hub.Run()
	expectedMessage := clientMessage{
		Type: "test",
	}
	hub.roomBroadcast <- newRoomBroadcast(client.RoomId, expectedMessage)

	got := <-clientChannel

	if !reflect.DeepEqual(expectedMessage, got) {
		t.Errorf("expected %v, got %v", expectedMessage, got)
	}
}

func TestHub_Run_WhenDeveloperGuessed(t *testing.T) {
	clientChannel := make(chan message)
	client := &Client{RoomId: "1", send: clientChannel, Role: Developer, Guess: 0}
	clients := make(map[*Client]bool)
	clients[client] = true

	hub := Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       clients,
		rooms:         make(map[string]bool),
	}
	go hub.Run()

	expectedMessage := newDeveloperGuessed()
	hub.roomBroadcast <- newRoomBroadcast(client.RoomId, expectedMessage)

	got := <-clientChannel

	if !reflect.DeepEqual(expectedMessage, got) {
		t.Errorf("expected %v, got %v", expectedMessage, got)
	}
}

func TestHub_Run_WhenEveryDeveloperInRoomGuessed(t *testing.T) {
	clientChannel := make(chan message)
	client := &Client{RoomId: "1", send: clientChannel, Role: Developer, Guess: 1}
	clients := make(map[*Client]bool)
	clients[client] = true

	hub := Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       clients,
		rooms:         make(map[string]bool),
	}
	go hub.Run()

	hub.roomBroadcast <- newRoomBroadcast(client.RoomId, newDeveloperGuessed())

	got := <-clientChannel

	switch got.(type) {
	case developerGuessed:
		t.Error("expected everyoneGuessed, got developerGuessed")
	}
}

func TestHub_RunWhenResetRound(t *testing.T) {
	roomId := "1"
	clientChannel := make(chan message)
	client := &Client{RoomId: roomId, send: clientChannel, Role: Developer, Guess: 1}
	clients := make(map[*Client]bool)
	clients[client] = true

	rooms := make(map[string]bool)
	rooms[roomId] = true

	hub := Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       clients,
		rooms:         rooms,
	}
	go hub.Run()

	expectedMessage := newResetRound()
	hub.roomBroadcast <- newRoomBroadcast(client.RoomId, expectedMessage)

	got := <-clientChannel

	if client.Guess > 0 {
		t.Error("expected client guess to reset to 0")
	}

	if _, ok := hub.rooms[roomId]; ok {
		t.Error("expected roomId to be deleted")
	}

	if !reflect.DeepEqual(expectedMessage, got) {
		t.Errorf("expected %v, got %v", expectedMessage, got)
	}
}

func TestHub_Run_ForEveryOtherMessage(t *testing.T) {
	clientChannel := make(chan message)
	client := &Client{RoomId: "1", send: clientChannel, Role: Developer, Guess: 1}
	clients := make(map[*Client]bool)
	clients[client] = true

	hub := Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       clients,
		rooms:         make(map[string]bool),
	}
	go hub.Run()

	expectedMessage := newJoin()
	hub.roomBroadcast <- newRoomBroadcast(client.RoomId, expectedMessage)

	got := <-clientChannel

	if !reflect.DeepEqual(expectedMessage, got) {
		t.Errorf("expected %v, got %v", expectedMessage, got)
	}
}
