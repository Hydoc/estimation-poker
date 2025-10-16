package internal

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Hydoc/guess-dev/backend/internal/assert"
	"github.com/google/uuid"
)

func TestNewRoom(t *testing.T) {
	expectedRoomId := RoomId("Test")
	room := NewRoom(expectedRoomId, make(chan<- RoomId), "")
	assert.Equal(t, room.Id, expectedRoomId)
	assert.False(t, room.InProgress)
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
					guess: 1,
					Role:  Developer,
				}: true,
				{
					Role: ProductOwner,
				}: true,
			},
		},
		{
			name: "everyone guessed because developer skipped",
			want: true,
			clients: map[*Client]bool{
				{
					guess:  0,
					doSkip: true,
					Role:   Developer,
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
					guess:  0,
					doSkip: false,
					Role:   Developer,
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
				Clients: test.clients,
			}
			got := room.everyDevIsDone()
			assert.Equal(t, got, test.want)
		})
	}
}

func TestRoom_Run_RegisteringAClient(t *testing.T) {
	room := NewRoom("Test", make(chan<- RoomId), "")
	client := &Client{}
	go room.Run()

	room.Join <- client

	room.clientMu.Lock()
	if _, ok := room.Clients[client]; !ok {
		t.Error("expected room to have client")
	}
	room.clientMu.Unlock()
}

func TestRoom_Run_DeletingAClientAndDestroyingTheRoom(t *testing.T) {
	destroyChannel := make(chan RoomId)
	roomId := RoomId("Test")
	client := &Client{}
	room := &Room{
		Id:         roomId,
		InProgress: false,
		leave:      make(chan *Client),
		Join:       nil,
		Clients: map[*Client]bool{
			client: true,
		},
		Broadcast: nil,
		destroy:   destroyChannel,
	}
	go room.Run()

	room.leave <- client

	gotId := <-destroyChannel

	assert.Equal(t, gotId, roomId)

	room.clientMu.Lock()
	if _, ok := room.Clients[client]; ok {
		t.Error("expected room not to have client")
	}
	room.clientMu.Unlock()
}

func TestRoom_Run_BroadcastEstimate(t *testing.T) {
	clientSendChannel := make(chan *Message)
	client := &Client{
		send: clientSendChannel,
	}
	room := &Room{
		Id:         "Test",
		InProgress: false,
		leave:      nil,
		Join:       nil,
		Clients: map[*Client]bool{
			client: true,
		},
		Broadcast: make(chan *Message),
		destroy:   nil,
	}
	go room.Run()

	msg := &Message{
		Type: estimate,
		Data: nil,
	}
	room.Broadcast <- msg

	gotClientMsg := <-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, msg)
	assert.True(t, room.InProgress)
}

func TestRoom_Run_BroadcastDeveloperGuessed_EveryDeveloperGuessed(t *testing.T) {
	clientSendChannel := make(chan *Message)
	client := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		guess: 1,
	}
	room := &Room{
		Id:         "Test",
		InProgress: false,
		leave:      nil,
		Join:       nil,
		Clients: map[*Client]bool{
			client: true,
		},
		Broadcast: make(chan *Message),
		destroy:   nil,
	}
	go room.Run()
	room.Broadcast <- newDeveloperGuessed()

	gotClientMsg := <-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, newEveryoneIsDone())
}

func TestRoom_Run_BroadcastDeveloperGuessed_NotEveryoneGuessed(t *testing.T) {
	room := NewRoom(RoomId("Test"), make(chan<- RoomId), "Tester")
	go room.Run()

	clientSendChannel := make(chan *Message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	room.Join <- client
	room.Join <- &Client{
		Role:  Developer,
		guess: 0,
		send:  clientSendChannel,
	}
	msg := newDeveloperGuessed()
	room.Broadcast <- msg

	select {
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout")
	case gotClientMsg := <-clientSendChannel:
		assert.DeepEqual(t, gotClientMsg, msg)
	}
}

func TestRoom_Run_BroadcastNewRound(t *testing.T) {
	clientSendChannel := make(chan *Message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	developerToReset := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		guess: 2,
	}
	room := &Room{
		Id:         "Test",
		InProgress: true,
		leave:      nil,
		Join:       nil,
		Clients: map[*Client]bool{
			client:           true,
			developerToReset: true,
		},
		Broadcast: make(chan *Message),
		destroy:   nil,
	}
	go room.Run()

	msg := newNewRound()
	room.Broadcast <- msg

	gotClientMsg := <-clientSendChannel
	<-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, msg)
	assert.False(t, room.InProgress)
	assert.Equal(t, developerToReset.guess, 0)
}

func TestRoom_Run_BroadcastLeaveWhenRoomInProgress(t *testing.T) {
	clientSendChannel := make(chan *Message)
	broadcastChannel := make(chan *Message)
	client := &Client{
		send: clientSendChannel,
		Role: ProductOwner,
	}
	developerToReset := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		guess: 2,
	}
	room := &Room{
		Id:         "Test",
		InProgress: true,
		leave:      nil,
		Join:       nil,
		Clients: map[*Client]bool{
			client:           true,
			developerToReset: true,
		},
		Broadcast: broadcastChannel,
		destroy:   nil,
	}
	go room.Run()

	msg := newLeave()
	room.Broadcast <- msg
	gotClientMsg := <-clientSendChannel
	<-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, newNewRound())
	assert.False(t, room.InProgress)
	assert.Equal(t, developerToReset.guess, 0)
}

func TestRoom_lock(t *testing.T) {
	key := uuid.New()
	room := &Room{
		Id:             "Test",
		InProgress:     true,
		leave:          nil,
		Join:           nil,
		Clients:        make(map[*Client]bool),
		Broadcast:      make(chan *Message),
		destroy:        nil,
		IsLocked:       false,
		NameOfCreator:  "Bla",
		Key:            key,
		HashedPassword: make([]byte, 0),
	}

	got := room.lock("Bla", "top secret", key.String())

	assert.True(t, got)
	assert.True(t, room.IsLocked)
	assert.False(t, len(room.HashedPassword) == 0)
}

func TestRoom_lock_WhenLockingFails(t *testing.T) {
	key := uuid.New()
	room := &Room{
		Id:             "Test",
		InProgress:     true,
		leave:          nil,
		Join:           nil,
		Clients:        make(map[*Client]bool),
		Broadcast:      make(chan *Message),
		destroy:        nil,
		IsLocked:       false,
		NameOfCreator:  "Bla",
		Key:            key,
		HashedPassword: make([]byte, 0),
	}

	got := room.lock("ABC", "top secret", key.String())

	assert.False(t, got)
}

func TestRoom_open_WhenUserNotCreator(t *testing.T) {
	id := uuid.New()
	room := &Room{
		Id:            "Test",
		InProgress:    false,
		NameOfCreator: "some user",
		IsLocked:      false,
		Key:           id,
	}
	got := room.open("invalid user", id.String())

	assert.False(t, got)
}

func TestRoom_open_WhenKeyIsWrong(t *testing.T) {
	room := &Room{
		Id:            "Test",
		InProgress:    false,
		NameOfCreator: "some user",
		IsLocked:      false,
		Key:           uuid.New(),
	}

	got := room.open("some user", "incorrect Key")

	assert.False(t, got)
}

func TestRoom_lock_WhenLockingFailsDueToHashingFails(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	key := uuid.New()
	room := &Room{
		Id:             "Test",
		InProgress:     true,
		leave:          nil,
		Join:           nil,
		Clients:        make(map[*Client]bool),
		Broadcast:      make(chan *Message),
		destroy:        nil,
		IsLocked:       false,
		NameOfCreator:  "Bla",
		Key:            key,
		HashedPassword: make([]byte, 0),
	}

	got := room.lock("ABC", strings.Repeat("bla", 90), key.String())
	wantedLog := "could not hash password"

	assert.False(t, got)
	assert.StringContains(t, logBuffer.String(), wantedLog)
}
