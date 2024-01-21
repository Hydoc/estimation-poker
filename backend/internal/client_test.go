package internal

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

var upgrader = &websocket.Upgrader{}

func TestClient_NewProductOwner(t *testing.T) {
	expectedName := "Test Person"
	expectedRoomId := "Test"
	expectedRole := ProductOwner
	client := newProductOwner(expectedRoomId, expectedName, &Hub{}, &websocket.Conn{})

	if expectedName != client.Name {
		t.Errorf("expected %v, got %v", expectedName, client.Name)
	}

	if expectedRoomId != client.RoomId {
		t.Errorf("expected %v, got %v", expectedRoomId, client.RoomId)
	}

	if expectedRole != client.Role {
		t.Errorf("expected role %v, got %v", expectedRole, client.Role)
	}

	expectedJsonRepresentation := userDTO{
		"name": expectedName,
		"role": expectedRole,
	}

	got := client.toJson()

	if !reflect.DeepEqual(expectedJsonRepresentation, got) {
		t.Errorf("expected %v, got %v", expectedJsonRepresentation, got)
	}
}

func TestClient_NewDeveloper(t *testing.T) {
	expectedName := "Test Person"
	expectedRoomId := "Test"
	expectedRole := Developer
	expectedGuess := 0
	client := newDeveloper(expectedRoomId, expectedName, &Hub{}, &websocket.Conn{})

	if expectedName != client.Name {
		t.Errorf("expected name %v, got %v", expectedName, client.Name)
	}

	if expectedRoomId != client.RoomId {
		t.Errorf("expected roomId %v, got %v", expectedRoomId, client.RoomId)
	}

	if expectedRole != client.Role {
		t.Errorf("expected role %v, got %v", expectedRole, client.Role)
	}

	if expectedGuess != client.Guess {
		t.Errorf("expected guess %v, got %v", expectedGuess, client.Guess)
	}

	expectedJsonRepresentation := userDTO{
		"name":  expectedName,
		"role":  expectedRole,
		"guess": expectedGuess,
	}

	got := client.toJson()

	if !reflect.DeepEqual(expectedJsonRepresentation, got) {
		t.Errorf("expected %v, got %v", expectedJsonRepresentation, got)
	}
}

func TestClient_Reset(t *testing.T) {
	client := newDeveloper("1", "Any", &Hub{}, &websocket.Conn{})
	client.Guess = 2
	client.reset()

	if client.Guess > 0 {
		t.Errorf("expected guess to be 0, got %v", client.Guess)
	}
}

func TestClient_WebsocketReaderUnregisteringWhenErrorOccurred(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	unregisterChannel := make(chan *Client)
	roomBroadcastChannel := make(chan roomBroadcastMessage)
	hub := &Hub{
		roomBroadcast: roomBroadcastChannel,
		register:      make(chan *Client),
		unregister:    unregisterChannel,
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client := newDeveloper("1", "Any", hub, connection)
	go client.websocketReader()

	// throws an error when trying to ReadJSON in client
	connection.WriteMessage(websocket.TextMessage, []byte("hello"))

	got := <-unregisterChannel

	if got != client {
		t.Errorf("expected client to unregister")
	}

	expectedRoomBroadcastMsg := newRoomBroadcast(client.RoomId, newLeave())
	gotRoomBroadcastMsg := <-roomBroadcastChannel

	if !reflect.DeepEqual(expectedRoomBroadcastMsg, gotRoomBroadcastMsg) {
		t.Errorf("want %v, got %v", expectedRoomBroadcastMsg, gotRoomBroadcastMsg)
	}

	wantedLog := "read: invalid character 'h' looking for beginning of value"

	if !strings.Contains(logBuffer.String(), wantedLog) {
		t.Errorf("expected to log %v", wantedLog)
	}
}

func TestClient_WebsocketReaderWhenAnyMessageOccurred(t *testing.T) {
	roomBroadcastChannel := make(chan roomBroadcastMessage)
	hub := &Hub{
		roomBroadcast: roomBroadcastChannel,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := newDeveloper("1", "Any", hub, connection)
	go client.websocketReader()
	expectedMsg := clientMessage{
		Type: "",
		Data: nil,
	}
	connection.WriteJSON(expectedMsg)

	got := <-roomBroadcastChannel

	if got.RoomId != client.RoomId {
		t.Errorf("expected room id %s, got %s", client.RoomId, got.RoomId)
	}

	if got.message != expectedMsg {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func TestClient_WebsocketReaderWhenGuessMessageOccurredWithClientDeveloper(t *testing.T) {
	roomBroadcastChannel := make(chan roomBroadcastMessage)
	hub := &Hub{
		roomBroadcast: roomBroadcastChannel,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	clientChannel := make(chan message)
	client := &Client{
		connection: connection,
		Role:       Developer,
		send:       clientChannel,
		Name:       "Test",
		hub:        hub,
		RoomId:     "1",
	}
	go client.websocketReader()

	connection.WriteJSON(clientMessage{
		Type: "guess",
		Data: 2,
	})

	got := <-roomBroadcastChannel

	if got.RoomId != client.RoomId {
		t.Errorf("expected room id %s, got %s", client.RoomId, got.RoomId)
	}

	expectedClientMsg := newYouGuessed(2)
	gotClientMsg := <-clientChannel

	if !reflect.DeepEqual(expectedClientMsg, gotClientMsg) {
		t.Errorf("expected client msg %v, got %v", expectedClientMsg, gotClientMsg)
	}

	if client.Guess != 2 {
		t.Errorf("expected client to have guess 2, got %v", client.Guess)
	}
}

func TestClient_WebsocketReaderWhenNewRondMessageOccurredWithClientProductOwner(t *testing.T) {
	roomBroadcastChannel := make(chan roomBroadcastMessage)
	hub := &Hub{
		roomBroadcast: roomBroadcastChannel,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := newProductOwner("1", "Test", hub, connection)
	go client.websocketReader()

	connection.WriteJSON(clientMessage{
		Type: newRound,
	})

	got := <-roomBroadcastChannel

	if got.RoomId != client.RoomId {
		t.Errorf("expected room id %s, got %s", client.RoomId, got.RoomId)
	}

	expectedMsg := newResetRound()

	if !reflect.DeepEqual(expectedMsg, got.message) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func TestClient_WebsocketWriter(t *testing.T) {
	roomBroadcastChannel := make(chan roomBroadcastMessage)
	hub := &Hub{
		roomBroadcast: roomBroadcastChannel,
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	clientChannel := make(chan message)
	client := &Client{
		connection: connection,
		Role:       Developer,
		send:       clientChannel,
		Name:       "Test",
		hub:        hub,
		RoomId:     "1",
	}
	go client.websocketWriter()
	go client.websocketReader()

	// due to the echo websocket it writes to itself
	expectedMsg := clientMessage{
		Type: "hello",
	}
	clientChannel <- expectedMsg

	got := <-roomBroadcastChannel

	if got.RoomId != client.RoomId {
		t.Errorf("expected room id %v, got %v", client.RoomId, got.RoomId)
	}

	if !reflect.DeepEqual(expectedMsg, got.message) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("could not upgrade connection: ", err)
		return
	}

	defer conn.Close()
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		err = conn.WriteMessage(mt, msg)
		if err != nil {
			return
		}
	}
}
