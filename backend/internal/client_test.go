package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Hydoc/go-message"

	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func TestClient_NewProductOwner(t *testing.T) {
	expectedName := "Test Person"
	expectedRole := ProductOwner
	client := NewClient(expectedName, expectedRole, &Room{}, &websocket.Conn{}, message.NewBus(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	assert.Equal(t, client.Name, expectedName)
	assert.Equal(t, client.Role, expectedRole)

	want, err := json.Marshal(map[string]string{
		"name": expectedName,
		"role": expectedRole,
	})
	assert.NilError(t, err)

	got, err := json.Marshal(client)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func TestClient_NewClient(t *testing.T) {
	expectedName := "Test Person"
	expectedRole := Developer
	expectedGuess := 0
	client := NewClient(expectedName, expectedRole, &Room{}, &websocket.Conn{}, message.NewBus(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	assert.Equal(t, client.Name, expectedName)
	assert.Equal(t, client.Role, expectedRole)
	assert.Equal(t, client.Guess(), expectedGuess)
	assert.False(t, client.doSkip)

	want := map[string]any{
		"name":   expectedName,
		"role":   expectedRole,
		"isDone": false,
	}

	got, err := json.Marshal(client)
	assert.NilError(t, err)

	var gotMap map[string]any
	err = json.Unmarshal(got, &gotMap)
	assert.NilError(t, err)
	assert.DeepEqual(t, gotMap, want)
}

func TestClient_Reset(t *testing.T) {
	client := NewClient("Any", Developer, &Room{}, &websocket.Conn{}, message.NewBus(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	client.guess = 2
	client.newRound()

	assert.Equal(t, client.Guess(), 0)
}

func TestClient_WebsocketReaderWhenGuessMessageOccurredWithClientDeveloper(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
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
	bus.Register(guess, handleGuess)
	clientChannel := make(chan *OutgoingWebsocketMessage)
	client := &Client{
		connection: connection,
		Role:       Developer,
		send:       clientChannel,
		Name:       "Test",
		room:       room,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, OutgoingWebsocketMessage{
		Type: guess,
		Data: 2,
	})

	firstBroadcastMsg := <-broadcastChannel
	secondBroadcastMsg := <-broadcastChannel

	expectedClientMsg := newOutgoingWebsocketMessage(youGuessed, 2)
	gotClientMsg := <-clientChannel

	assert.DeepEqual(t, firstBroadcastMsg, newOutgoingWebsocketMessage(developerAction, nil))
	assert.DeepEqual(t, secondBroadcastMsg, newUsers(room.Clients))
	assert.DeepEqual(t, gotClientMsg, expectedClientMsg)
	assert.Equal(t, client.Guess(), 2)
}

func TestClient_websocketReaderRevealMessage(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
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
	bus.Register(reveal, handleReveal)
	client := NewClient("Test", ProductOwner, room, connection, bus, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	go client.WebsocketReader()
	go client.WebsocketWriter()
	expectedMessage := &OutgoingWebsocketMessage{
		Type: reveal,
		Data: []map[string]any{},
	}
	client.send <- newReveal(room.Clients)
	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMessage)
}

func TestClient_WebsocketReaderAddIssueMessage(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	room := &Room{
		broadcast:      broadcastChannel,
		HashedPassword: make([]byte, 0),
		GuessConfig: &GuessConfig{
			Guesses: make([]GuessConfigEntry, 0),
		},
		join:    make(chan *Client),
		leave:   make(chan *Client),
		Clients: make(map[*Client]bool),
	}

	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	connection, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	bus := message.NewBus()
	bus.Register(addIssue, handleAddIssue)
	client := NewClient("Test", ProductOwner, room, connection, bus, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	go client.WebsocketReader()
	go client.WebsocketWriter()
	expectedMessage := &OutgoingWebsocketMessage{
		Type: issues,
	}
	client.send <- &OutgoingWebsocketMessage{
		Type: addIssue,
		Data: "Issue to add",
	}
	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMessage)
	assert.DeepEqual(t, room.State().Issues, []*Issue{
		{
			Title: "Issue to add",
			Guess: -1,
		},
	})
}

func TestClient_WebsocketReaderWhenNewRoundMessageOccurredWithClientProductOwner(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
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
	bus.Register(newRound, handleNewRound)
	client := NewClient("Test", ProductOwner, room, connection, bus, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	go client.WebsocketReader()

	expectedMsg := newOutgoingWebsocketMessage(newRound, nil)
	wsjson.Write(context.Background(), connection, expectedMsg)

	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMsg)
}

func TestClient_WebsocketReader_WhenSkipRoundMessageOccurredWithClientDeveloper(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
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
	bus.Register(skipRound, handleSkipRound)
	clientChannel := make(chan *OutgoingWebsocketMessage)
	client := &Client{
		connection: connection,
		room:       room,
		Name:       "Test",
		Role:       Developer,
		send:       clientChannel,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, OutgoingWebsocketMessage{
		Type: skipRound,
	})

	expectedMsg := newOutgoingWebsocketMessage(developerAction, nil)
	secondExpectedMsg := newUsers(room.Clients)
	expectedClientMsg := newOutgoingWebsocketMessage(youSkipped, nil)
	firstBroadcast := <-broadcastChannel
	secondBroadcast := <-broadcastChannel
	gotClientMessage := <-clientChannel

	assert.DeepEqual(t, firstBroadcast, expectedMsg)
	assert.DeepEqual(t, secondBroadcast, secondExpectedMsg)
	assert.DeepEqual(t, gotClientMessage, expectedClientMsg)
}

func TestClient_WebsocketReader_WhenLockRoomMessageOccurredAnyClientCanLock(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	id := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("my cool pw"), bcrypt.DefaultCost)
	nameOfCreator := "Test"
	room := &Room{
		broadcast:      broadcastChannel,
		join:           make(chan *Client),
		leave:          make(chan *Client),
		Clients:        make(map[*Client]bool),
		NameOfCreator:  nameOfCreator,
		key:            id,
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
	bus.Register(lockRoom, handleLockRoom)
	client := &Client{
		connection: connection,
		room:       room,
		Name:       nameOfCreator,
		Role:       Developer,
		send:       nil,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, OutgoingWebsocketMessage{
		Type: lockRoom,
		Data: map[string]any{
			"password": "my cool pw",
			"key":      id.String(),
		},
	})

	expectedMsg := newOutgoingWebsocketMessage(roomLocked, nil)
	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMsg)
}

func TestClient_WebsocketReader_WhenOpenRoomMessageOccurred(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	id := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("my cool pw"), bcrypt.DefaultCost)
	room := &Room{
		broadcast:      broadcastChannel,
		join:           make(chan *Client),
		leave:          make(chan *Client),
		Clients:        make(map[*Client]bool),
		NameOfCreator:  "Test",
		key:            id,
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
	bus.Register(openRoom, handleOpenRoom)
	client := &Client{
		connection: connection,
		room:       room,
		Name:       "Test",
		Role:       Developer,
		send:       nil,
		bus:        bus,
	}
	go client.WebsocketReader()

	wsjson.Write(context.Background(), connection, OutgoingWebsocketMessage{
		Type: openRoom,
		Data: map[string]any{
			"key": id.String(),
		},
	})

	expectedMsg := newOutgoingWebsocketMessage(roomOpened, nil)
	got := <-broadcastChannel

	assert.DeepEqual(t, got, expectedMsg)
}

func TestClient_WebsocketWriter(t *testing.T) {
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
	room := &Room{
		broadcast: broadcastChannel,
		join:      make(chan *Client),
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

	clientChannel := make(chan *OutgoingWebsocketMessage)
	bus := message.NewBus()
	bus.Register(estimate, handleEstimate)
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
	expectedMsg := &OutgoingWebsocketMessage{
		Type: estimate,
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
