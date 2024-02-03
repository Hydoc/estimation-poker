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
	"sync"
	"testing"
)

var upgrader = &websocket.Upgrader{}

type failingMessage struct{}

func (msg failingMessage) ToJson() messageDTO {
	return map[string]interface{}{
		"b": make(chan int),
	}
}

func TestClient_NewProductOwner(t *testing.T) {
	expectedName := "Test Person"
	expectedRole := ProductOwner
	client := newClient(expectedName, expectedRole, &Room{}, &websocket.Conn{})

	if expectedName != client.Name {
		t.Errorf("expected %v, got %v", expectedName, client.Name)
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
	expectedRole := Developer
	expectedGuess := 0
	client := newClient(expectedName, expectedRole, &Room{}, &websocket.Conn{})

	if expectedName != client.Name {
		t.Errorf("expected name %v, got %v", expectedName, client.Name)
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
	client := newClient("Any", Developer, &Room{}, &websocket.Conn{})
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
	broadcastChannel := make(chan message)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
		leave:     unregisterChannel,
		clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	client := newClient("Any", Developer, room, connection)
	go client.websocketReader()

	// throws an error when trying to ReadJSON in client
	connection.WriteMessage(websocket.TextMessage, []byte("hello"))

	got := <-unregisterChannel

	if got != client {
		t.Errorf("expected client to unregister")
	}

	expectedRoomBroadcastMsg := newLeave()
	gotRoomBroadcastMsg := <-broadcastChannel

	if !reflect.DeepEqual(expectedRoomBroadcastMsg, gotRoomBroadcastMsg) {
		t.Errorf("want %v, got %v", expectedRoomBroadcastMsg, gotRoomBroadcastMsg)
	}

	wantedLog := "read: invalid character 'h' looking for beginning of value"

	if !strings.Contains(logBuffer.String(), wantedLog) {
		t.Errorf("expected to log %v", wantedLog)
	}
}

func TestClient_WebsocketReaderWhenAnyMessageOccurred(t *testing.T) {
	broadcastChannel := make(chan message)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := newClient("Any", Developer, room, connection)
	go client.websocketReader()
	expectedMsg := clientMessage{
		Type: "",
		Data: nil,
	}
	connection.WriteJSON(expectedMsg)

	got := <-broadcastChannel

	if got != expectedMsg {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func TestClient_WebsocketReaderWhenGuessMessageOccurredWithClientDeveloper(t *testing.T) {
	broadcastChannel := make(chan message)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
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
		room:       room,
	}
	go client.websocketReader()

	connection.WriteJSON(clientMessage{
		Type: "guess",
		Data: 2,
	})

	<-broadcastChannel

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
	broadcastChannel := make(chan message)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := newClient("Test", ProductOwner, room, connection)
	go client.websocketReader()

	connection.WriteJSON(clientMessage{
		Type: newRound,
	})

	expectedMsg := newResetRound()
	got := <-broadcastChannel

	if !reflect.DeepEqual(expectedMsg, got) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func TestClient_WebsocketWriter(t *testing.T) {
	broadcastChannel := make(chan message)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
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
		room:       room,
	}
	go client.websocketWriter()
	go client.websocketReader()

	// due to the echo websocket it writes to itself
	expectedMsg := clientMessage{
		Type: "hello",
	}
	clientChannel <- expectedMsg

	got := <-broadcastChannel

	if !reflect.DeepEqual(expectedMsg, got) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func TestClient_WebsocketWriter_WhenErrorOccurred(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	broadcastChannel := make(chan message)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var wg sync.WaitGroup
	clientChannel := make(chan message)
	client := &Client{
		connection: connection,
		Role:       Developer,
		send:       clientChannel,
		Name:       "Test",
		room:       room,
	}
	go func() {
		wg.Add(1)
		defer wg.Done()
		client.websocketWriter()
	}()

	clientChannel <- failingMessage{}

	wantLog := "unsupported type: chan int"

	wg.Wait()

	if !strings.Contains(logBuffer.String(), wantLog) {
		t.Errorf("expected to log %v, got %#v", wantLog, logBuffer.String())
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
