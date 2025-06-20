package internal

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewRoom(t *testing.T) {
	expectedRoomId := RoomId("Test")
	room := newRoom(expectedRoomId, make(chan<- RoomId), "")

	if room.id != expectedRoomId {
		t.Errorf("want room id %v, got %v", expectedRoomId, room.id)
	}

	if room.inProgress {
		t.Error("expected room not to be in progress")
	}
}

func TestRoom_everyDevGuessed(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		clients map[*Client]bool
	}{
		{
			name: "everyone guessed",
			want: true,
			clients: map[*Client]bool{
				{
					Guess: 1,
					Role:  Developer,
				}: true,
				{
					Role: ProductOwner,
				}: true,
			},
		},
		{
			name: "not everyone guessed",
			want: false,
			clients: map[*Client]bool{
				{
					Guess: 0,
					Role:  Developer,
				}: true,
				{
					Role: ProductOwner,
				}: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			room := &Room{
				clients: test.clients,
			}
			got := room.everyDevIsDone()
			if got != test.want {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func TestRoom_Run_RegisteringAClient(t *testing.T) {
	room := newRoom("Test", make(chan<- RoomId), "")
	client := &Client{}
	go room.Run()

	room.join <- client

	room.clientMu.Lock()
	if _, ok := room.clients[client]; !ok {
		t.Error("expected room to have client")
	}
	room.clientMu.Unlock()
}

func TestRoom_Run_DeletingAClientAndDestroyingTheRoom(t *testing.T) {
	destroyChannel := make(chan RoomId)
	roomId := RoomId("Test")
	client := &Client{}
	room := &Room{
		id:         roomId,
		inProgress: false,
		leave:      make(chan *Client),
		join:       nil,
		clients: map[*Client]bool{
			client: true,
		},
		broadcast: nil,
		destroy:   destroyChannel,
	}
	go room.Run()

	room.leave <- client

	gotId := <-destroyChannel
	if gotId != roomId {
		t.Errorf("want room id %v, got %v", roomId, gotId)
	}

	room.clientMu.Lock()
	if _, ok := room.clients[client]; ok {
		t.Error("expected room not to have client")
	}
	room.clientMu.Unlock()
}

func TestRoom_Run_BroadcastEstimate(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send: clientSendChannel,
	}
	room := &Room{
		id:         "Test",
		inProgress: false,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()

	msg := message{
		Type: estimate,
		Data: nil,
	}
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, msg) {
		t.Errorf("want message %v, got %v", msg, gotClientMsg)
	}

	if !room.inProgress {
		t.Error("expected room to be in progress")
	}
}

func TestRoom_Run_BroadcastDeveloperGuessed_EveryDeveloperGuessed(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		Guess: 1,
	}
	room := &Room{
		id:         "Test",
		inProgress: false,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()
	room.broadcast <- newDeveloperGuessed()

	gotClientMsg := <-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, newEveryoneIsDone()) {
		t.Errorf("want msg %v, got %v", newEveryoneIsDone(), gotClientMsg)
	}
}

func TestRoom_Run_BroadcastDeveloperGuessed_NotEveryoneGuessed(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	room := &Room{
		id:         "Test",
		inProgress: false,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client: true,
			{
				Role:  Developer,
				Guess: 0,
			}: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()
	msg := newDeveloperGuessed()
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, msg) {
		t.Errorf("want msg %v, got %v", msg, gotClientMsg)
	}
}

func TestRoom_Run_BroadcastResetRound(t *testing.T) {
	clientSendChannel := make(chan message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	developerToReset := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		Guess: 2,
	}
	room := &Room{
		id:         "Test",
		inProgress: true,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client:           true,
			developerToReset: true,
		},
		broadcast: make(chan message),
		destroy:   nil,
	}
	go room.Run()

	msg := newResetRound()
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel
	<-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, msg) {
		t.Errorf("want msg %v, got %v", msg, gotClientMsg)
	}

	if room.inProgress {
		t.Error("expected room not to be in progress")
	}

	if developerToReset.Guess > 0 {
		t.Error("expected developer to be resetted")
	}
}

func TestRoom_Run_BroadcastLeaveWhenRoomInProgress(t *testing.T) {
	clientSendChannel := make(chan message)
	broadcastChannel := make(chan message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	developerToReset := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		Guess: 2,
	}
	room := &Room{
		id:         "Test",
		inProgress: true,
		leave:      nil,
		join:       nil,
		clients: map[*Client]bool{
			client:           true,
			developerToReset: true,
		},
		broadcast: broadcastChannel,
		destroy:   nil,
	}
	go room.Run()

	msg := newLeave()
	room.broadcast <- msg
	gotClientMsg := <-clientSendChannel
	<-clientSendChannel

	if !reflect.DeepEqual(gotClientMsg, newResetRound()) {
		t.Errorf("want msg %v, got %v", newResetRound(), gotClientMsg)
	}

	if room.inProgress {
		t.Error("expected room not to be in progress")
	}

	if developerToReset.Guess > 0 {
		t.Error("expected developer to be resetted")
	}
}

func TestRoom_lock(t *testing.T) {
	key := uuid.New()
	room := &Room{
		id:             "Test",
		inProgress:     true,
		leave:          nil,
		join:           nil,
		clients:        make(map[*Client]bool),
		broadcast:      make(chan message),
		destroy:        nil,
		isLocked:       false,
		nameOfCreator:  "Bla",
		key:            key,
		hashedPassword: make([]byte, 0),
	}

	got := room.lock("Bla", "top secret", key.String())

	if got != true {
		t.Errorf("got %v, want true", got)
	}

	if !room.isLocked {
		t.Errorf("wanted room to be locked")
	}

	if len(room.hashedPassword) == 0 {
		t.Errorf("wanted room to have hashed password")
	}
}

func TestRoom_lock_WhenLockingFails(t *testing.T) {
	key := uuid.New()
	room := &Room{
		id:             "Test",
		inProgress:     true,
		leave:          nil,
		join:           nil,
		clients:        make(map[*Client]bool),
		broadcast:      make(chan message),
		destroy:        nil,
		isLocked:       false,
		nameOfCreator:  "Bla",
		key:            key,
		hashedPassword: make([]byte, 0),
	}

	got := room.lock("ABC", "top secret", key.String())

	if got != false {
		t.Errorf("got %v, want false", got)
	}
}

func TestRoom_open_WhenUserNotCreator(t *testing.T) {
	id := uuid.New()
	room := &Room{
		id:            "Test",
		inProgress:    false,
		nameOfCreator: "some user",
		isLocked:      false,
		key:           id,
	}

	if got := room.open("invalid user", id.String()); got != false {
		t.Error("expected to be false")
	}
}

func TestRoom_open_WhenKeyIsWrong(t *testing.T) {
	room := &Room{
		id:            "Test",
		inProgress:    false,
		nameOfCreator: "some user",
		isLocked:      false,
		key:           uuid.New(),
	}

	if got := room.open("some user", "incorrect key"); got != false {
		t.Error("expected to be false")
	}
}

func TestRoom_lock_WhenLockingFailsDueToHashingFails(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	key := uuid.New()
	room := &Room{
		id:             "Test",
		inProgress:     true,
		leave:          nil,
		join:           nil,
		clients:        make(map[*Client]bool),
		broadcast:      make(chan message),
		destroy:        nil,
		isLocked:       false,
		nameOfCreator:  "Bla",
		key:            key,
		hashedPassword: make([]byte, 0),
	}

	got := room.lock("ABC", strings.Repeat("bla", 90), key.String())
	wantedLog := "could not hash password"

	if got != false {
		t.Errorf("got %v, want false", got)
	}

	if !strings.Contains(logBuffer.String(), wantedLog) {
		t.Errorf("expected to log %s", wantedLog)
	}
}
