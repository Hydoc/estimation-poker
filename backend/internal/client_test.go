package internal

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Hydoc/go-message"
	"github.com/Hydoc/guess-dev/backend/internal/assert"
)

func TestClient_NewProductOwner(t *testing.T) {
	expectedName := "Test Person"
	expectedRole := ProductOwner
	client := NewClient(expectedName, expectedRole, &Room{}, &websocket.Conn{}, message.NewBus())

	if expectedName != client.Name {
		t.Errorf("expected %v, got %v", expectedName, client.Name)
	}

	if expectedRole != client.Role {
		t.Errorf("expected role %v, got %v", expectedRole, client.Role)
	}

	expectedJsonRepresentation := UserDTO{
		"name": expectedName,
		"role": expectedRole,
	}

	got := client.ToJson()

	assert.DeepEqual(t, got, expectedJsonRepresentation)
}

func TestClient_NewClient(t *testing.T) {
	expectedName := "Test Person"
	expectedRole := Developer
	expectedGuess := 0
	client := NewClient(expectedName, expectedRole, &Room{}, &websocket.Conn{}, message.NewBus())

	assert.Equal(t, client.Name, expectedName)
	assert.Equal(t, client.Role, expectedRole)
	assert.Equal(t, client.guess, expectedGuess)
	assert.False(t, client.doSkip)

	expectedJsonRepresentation := UserDTO{
		"name":   expectedName,
		"role":   expectedRole,
		"isDone": false,
	}

	got := client.ToJson()

	assert.DeepEqual(t, got, expectedJsonRepresentation)
}

func TestClient_Reset(t *testing.T) {
	client := NewClient("Any", Developer, &Room{}, &websocket.Conn{}, message.NewBus())
	client.guess = 2
	client.newRound()

	if client.guess > 0 {
		t.Errorf("expected guess to be 0, got %v", client.guess)
	}
}

func TestClient_WebsocketReaderWhenGuessMessageOccurredWithClientDeveloper(t *testing.T) {
	broadcastChannel := make(chan *Message)
	room := &Room{
		Broadcast: broadcastChannel,
		Join:      make(chan *Client),
		leave:     make(chan *Client),
		Clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	bus := message.NewBus()
	bus.Register(Guess, HandleGuess)
	clientChannel := make(chan *Message)
	client := &Client{
		connection: connection,
		Role:       Developer,
		send:       clientChannel,
		Name:       "Test",
		room:       room,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, Message{
		Type: Guess,
		Data: 2,
	})

	<-broadcastChannel

	expectedClientMsg := newYouGuessed(2)
	gotClientMsg := <-clientChannel

	assert.DeepEqual(t, gotClientMsg, expectedClientMsg)
	assert.Equal(t, client.guess, 2)
}

func TestClient_websocketReaderRevealMessage(t *testing.T) {
	broadcastChannel := make(chan *Message)
	room := &Room{
		Broadcast: broadcastChannel,
		Join:      make(chan *Client),
		leave:     make(chan *Client),
		Clients:   make(map[*Client]bool),
	}

	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	bus := message.NewBus()
	bus.Register(Reveal, HandleReveal)
	client := NewClient("Test", ProductOwner, room, connection, bus)
	go client.WebsocketReader()
	go client.WebsocketWriter()
	expectedMessage := &Message{
		Type: Reveal,
		Data: []map[string]any{},
	}
	client.send <- newReveal(room.Clients)
	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMessage)
}

func TestClient_WebsocketReaderWhenNewRoundMessageOccurredWithClientProductOwner(t *testing.T) {
	broadcastChannel := make(chan *Message)
	room := &Room{
		Broadcast: broadcastChannel,
		Join:      make(chan *Client),
		leave:     make(chan *Client),
		Clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	bus := message.NewBus()
	bus.Register(NewRound, HandleNewRound)
	client := NewClient("Test", ProductOwner, room, connection, bus)
	go client.WebsocketReader()

	expectedMsg := newNewRound()
	wsjson.Write(context.Background(), connection, expectedMsg)

	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMsg)
}

func TestClient_WebsocketReader_WhenSkipRoundMessageOccurredWithClientDeveloper(t *testing.T) {
	broadcastChannel := make(chan *Message)
	room := &Room{
		Broadcast: broadcastChannel,
		Join:      make(chan *Client),
		leave:     make(chan *Client),
		Clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	bus := message.NewBus()
	bus.Register(SkipRound, HandleSkipRound)
	clientChannel := make(chan *Message)
	client := &Client{
		connection: connection,
		room:       room,
		Name:       "Test",
		Role:       Developer,
		send:       clientChannel,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, Message{
		Type: SkipRound,
	})

	expectedMsg := newDeveloperSkipped()
	expectedClientMsg := newYouSkipped()
	got := <-broadcastChannel
	gotClientMessage := <-clientChannel

	assert.DeepEqual(t, got, expectedMsg)
	assert.DeepEqual(t, gotClientMessage, expectedClientMsg)
}

func TestClient_WebsocketReader_WhenLockRoomMessageOccurredAnyClientCanLock(t *testing.T) {
	broadcastChannel := make(chan *Message)
	id := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("my cool pw"), bcrypt.DefaultCost)
	nameOfCreator := "Test"
	room := &Room{
		Broadcast:      broadcastChannel,
		Join:           make(chan *Client),
		leave:          make(chan *Client),
		Clients:        make(map[*Client]bool),
		NameOfCreator:  nameOfCreator,
		Key:            id,
		HashedPassword: hashedPassword,
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	bus := message.NewBus()
	bus.Register(LockRoom, HandleLockRoom)
	client := &Client{
		connection: connection,
		room:       room,
		Name:       nameOfCreator,
		Role:       Developer,
		send:       nil,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, Message{
		Type: LockRoom,
		Data: map[string]any{
			"password": "my cool pw",
			"key":      id.String(),
		},
	})

	expectedMsg := newRoomLocked()
	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMsg)
}

func TestClient_WebsocketReader_WhenOpenRoomMessageOccurred(t *testing.T) {
	broadcastChannel := make(chan *Message)
	id := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("my cool pw"), bcrypt.DefaultCost)
	room := &Room{
		Broadcast:      broadcastChannel,
		Join:           make(chan *Client),
		leave:          make(chan *Client),
		Clients:        make(map[*Client]bool),
		NameOfCreator:  "Test",
		Key:            id,
		HashedPassword: hashedPassword,
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	bus := message.NewBus()
	bus.Register(OpenRoom, HandleOpenRoom)
	client := &Client{
		connection: connection,
		room:       room,
		Name:       "Test",
		Role:       Developer,
		send:       nil,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, Message{
		Type: OpenRoom,
		Data: map[string]any{
			"key": id.String(),
		},
	})

	expectedMsg := newRoomOpened()
	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMsg)
}

func TestClient_WebsocketWriter(t *testing.T) {
	broadcastChannel := make(chan *Message)
	room := &Room{
		Broadcast: broadcastChannel,
		Join:      make(chan *Client),
		leave:     make(chan *Client),
		Clients:   make(map[*Client]bool),
	}
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	clientChannel := make(chan *Message)
	bus := message.NewBus()
	bus.Register(Estimate, HandleEstimate)
	client := &Client{
		connection: connection,
		Role:       ProductOwner,
		send:       clientChannel,
		Name:       "Test",
		room:       room,
		bus:        bus,
	}
	go client.WebsocketWriter()
	go client.WebsocketReader()

	// due to the echo websocket it writes to itself
	expectedMsg := &Message{
		Type: Estimate,
		Data: "a-ticket",
	}
	clientChannel <- expectedMsg

	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMsg)
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
