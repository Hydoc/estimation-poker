package internal

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type failingMessage struct{}

func (msg failingMessage) ToJson() messageDTO {
	return map[string]any{
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

func TestClient_NewClient(t *testing.T) {
	expectedName := "Test Person"
	expectedRole := Developer
	expectedGuess := 0
	expectedDoSkip := false
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

	if expectedDoSkip != client.DoSkip {
		t.Errorf("expected do skip to be false, got %v", client.DoSkip)
	}

	expectedJsonRepresentation := userDTO{
		"name":   expectedName,
		"role":   expectedRole,
		"isDone": false,
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

func TestClient_WebsocketReaderWhenAnyMessageOccurred(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	room := &Room{
		broadcast: make(chan message),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := newClient("Any", Developer, room, connection)
	go client.websocketReader()

	unknownMessage := message{
		Type: "",
		Data: nil,
	}

	wsjson.Write(context.Background(), connection, unknownMessage)

	want := fmt.Sprintf("unknown message %#v", unknownMessage)

	if !strings.Contains(logBuffer.String(), want) {
		t.Errorf("expected %v, got %v", want, logBuffer.String())
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

	connection, _, err := websocket.Dial(context.Background(), url, nil)
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

	wsjson.Write(context.Background(), connection, message{
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

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := newClient("Test", ProductOwner, room, connection)
	go client.websocketReader()

	wsjson.Write(context.Background(), connection, message{
		Type: newRound,
	})

	expectedMsg := newResetRound()
	got := <-broadcastChannel

	if !reflect.DeepEqual(expectedMsg, got) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func TestClient_WebsocketReader_WhenSkipRoundMessageOccurredWithClientDeveloper(t *testing.T) {
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

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	clientChannel := make(chan message)
	client := &Client{
		connection: connection,
		room:       room,
		Name:       "Test",
		Role:       Developer,
		send:       clientChannel,
	}
	go client.websocketReader()

	wsjson.Write(context.Background(), connection, message{
		Type: skipRound,
	})

	expectedMsg := newSkipRound()
	expectedClientMsg := newYouSkipped()
	got := <-broadcastChannel
	gotClientMessage := <-clientChannel

	if !reflect.DeepEqual(expectedMsg, got) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
	if !reflect.DeepEqual(expectedClientMsg, gotClientMessage) {
		t.Errorf("expected %v, got %v", expectedClientMsg, gotClientMessage)
	}
}

func TestClient_WebsocketReader_WhenLockRoomMessageOccurredAnyClientCanLock(t *testing.T) {
	broadcastChannel := make(chan message)
	id := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("my cool pw"), bcrypt.DefaultCost)
	room := &Room{
		broadcast:      broadcastChannel,
		join:           make(chan *Client),
		leave:          make(chan *Client),
		clients:        make(map[*Client]bool),
		nameOfCreator:  "Test",
		key:            id,
		hashedPassword: hashedPassword,
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &Client{
		connection: connection,
		room:       room,
		Name:       "Test",
		Role:       Developer,
		send:       nil,
	}
	go client.websocketReader()

	wsjson.Write(context.Background(), connection, message{
		Type: lockRoom,
		Data: map[string]any{
			"password": "my cool pw",
			"key":      id.String(),
		},
	})

	expectedMsg := newRoomLocked()
	got := <-broadcastChannel

	if !reflect.DeepEqual(expectedMsg, got) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func TestClient_WebsocketReader_WhenOpenRoomMessageOccurred(t *testing.T) {
	broadcastChannel := make(chan message)
	id := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("my cool pw"), bcrypt.DefaultCost)
	room := &Room{
		broadcast:      broadcastChannel,
		join:           make(chan *Client),
		leave:          make(chan *Client),
		clients:        make(map[*Client]bool),
		nameOfCreator:  "Test",
		key:            id,
		hashedPassword: hashedPassword,
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	client := &Client{
		connection: connection,
		room:       room,
		Name:       "Test",
		Role:       Developer,
		send:       nil,
	}
	go client.websocketReader()

	wsjson.Write(context.Background(), connection, message{
		Type: openRoom,
		Data: map[string]any{
			"key": id.String(),
		},
	})

	expectedMsg := newRoomOpened()
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

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	clientChannel := make(chan message)
	client := &Client{
		connection: connection,
		Role:       ProductOwner,
		send:       clientChannel,
		Name:       "Test",
		room:       room,
	}
	go client.websocketWriter()
	go client.websocketReader()

	// due to the echo websocket it writes to itself
	expectedMsg := message{
		Type: estimate,
	}
	clientChannel <- expectedMsg

	got := <-broadcastChannel

	if !reflect.DeepEqual(expectedMsg, got) {
		t.Errorf("expected %v, got %v", expectedMsg, got)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("could not upgrade connection: ", err)
		return
	}

	defer conn.CloseNow()
	for {
		var data any
		err = wsjson.Read(context.Background(), conn, &data)
		if err != nil {
			return
		}

		err = wsjson.Write(context.Background(), conn, data)
		if err != nil {
			return
		}
	}
}
