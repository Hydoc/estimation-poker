package internal

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func TestNewRoom(t *testing.T) {
	expectedRoomId := uuid.New()
	room := NewRoom(expectedRoomId, make(chan<- uuid.UUID), "", slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), new(GuessConfig))
	assert.Equal(t, room.Id, expectedRoomId)
	assert.False(t, room.inProgress)
}

func TestRoom_ConnectionState(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		room     func() *Room
		want     ConnectionState
	}{
		{
			name:     "room in progress",
			username: "Test",
			password: "",
			room: func() *Room {
				return &Room{
					inProgress: true,
				}
			},
			want: ConnectionState{
				CanConnect: false,
				Reason:     ErrRoundStarted.Error(),
			},
		},
		{
			name:     "locked but password incorrect",
			username: "Test",
			password: "incorrect",
			room: func() *Room {
				hashed, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatal(err)
				}
				return &Room{
					inProgress:     false,
					HashedPassword: hashed,
				}
			},
			want: ConnectionState{
				CanConnect: false,
				Reason:     ErrWrongPassword.Error(),
			},
		},
		{
			name:     "username already taken",
			username: "Test",
			password: "",
			room: func() *Room {
				return &Room{
					inProgress:     false,
					HashedPassword: make([]byte, 0),
					Clients: map[*Client]bool{
						&Client{
							Name: "Test",
						}: true,
					},
				}
			},
			want: ConnectionState{
				CanConnect: false,
				Reason:     ErrUsernameTaken.Error(),
			},
		},
		{
			name:     "can connect",
			username: "Test",
			password: "",
			room: func() *Room {
				return &Room{
					inProgress:     false,
					HashedPassword: make([]byte, 0),
					Clients:        make(map[*Client]bool),
				}
			},
			want: ConnectionState{
				CanConnect: true,
				Reason:     "",
			},
		},
		{
			name:     "can connect through password validation",
			username: "Test",
			password: "correct",
			room: func() *Room {
				hashed, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatal(err)
				}
				return &Room{
					inProgress:     false,
					HashedPassword: hashed,
				}
			},
			want: ConnectionState{
				CanConnect: true,
				Reason:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.room().ConnectionState(tt.username, tt.password)

			assert.DeepEqual(t, got, tt.want)
		})
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
	room := NewRoom(uuid.New(), make(chan<- uuid.UUID), "", slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), new(GuessConfig))
	client := &Client{}
	go room.Run()

	room.join <- client

	room.clientMu.Lock()
	if _, ok := room.Clients[client]; !ok {
		t.Error("expected room to have client")
	}
	room.clientMu.Unlock()
}

func TestRoom_Run_DeletingAClientAndDestroyingTheRoom(t *testing.T) {
	destroyChannel := make(chan uuid.UUID)
	roomId := uuid.New()
	client := &Client{}
	room := &Room{
		Id:         roomId,
		inProgress: false,
		leave:      make(chan *Client),
		join:       nil,
		Clients: map[*Client]bool{
			client: true,
		},
		broadcast: nil,
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
	clientSendChannel := make(chan *OutgoingWebsocketMessage)
	client := &Client{
		send: clientSendChannel,
	}
	room := &Room{
		Id:         uuid.New(),
		inProgress: false,
		leave:      nil,
		join:       nil,
		Clients: map[*Client]bool{
			client: true,
		},
		broadcast: make(chan *OutgoingWebsocketMessage),
		destroy:   nil,
	}
	go room.Run()

	msg := &OutgoingWebsocketMessage{
		Type: estimate,
		Data: nil,
	}
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, msg)
	assert.True(t, room.inProgress)
}

func TestRoom_Run_BroadcastDeveloperGuessed_EveryDeveloperGuessed(t *testing.T) {
	clientSendChannel := make(chan *OutgoingWebsocketMessage)
	client := &Client{
		send:  clientSendChannel,
		Role:  Developer,
		guess: 1,
	}
	room := &Room{
		Id:         uuid.New(),
		inProgress: false,
		leave:      nil,
		join:       nil,
		Clients: map[*Client]bool{
			client: true,
		},
		broadcast: make(chan *OutgoingWebsocketMessage),
		destroy:   nil,
	}
	go room.Run()
	room.broadcast <- newOutgoingWebsocketMessage(developerAction, nil)

	gotClientMsg := <-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, newOutgoingWebsocketMessage(everyoneDone, nil))
}

func TestRoom_Run_BroadcastDeveloperGuessed_NotEveryoneGuessed(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	room := NewRoom(uuid.New(), make(chan<- uuid.UUID), "Tester", logger, new(GuessConfig))
	go room.Run()

	clientSendChannel := make(chan *OutgoingWebsocketMessage)
	client := &Client{
		Name: "A",
		send: clientSendChannel,
		Role: ProductOwner,
	}
	room.join <- client
	room.join <- &Client{
		Name:  "B",
		Role:  Developer,
		guess: 0,
		send:  clientSendChannel,
	}
	msg := newOutgoingWebsocketMessage(developerAction, nil)
	room.broadcast <- msg

	gotClientMsg := <-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, newUsers(room.Clients))
}

func TestRoom_Run_BroadcastNewRound(t *testing.T) {
	clientSendChannel := make(chan *OutgoingWebsocketMessage)
	client := &Client{
		Name: "do nothing",
		send: clientSendChannel,
		Role: ProductOwner,
	}
	developerToReset := &Client{
		Name:  "reset me",
		send:  clientSendChannel,
		Role:  Developer,
		guess: 2,
	}
	room := &Room{
		Id:         uuid.New(),
		inProgress: true,
		leave:      nil,
		join:       nil,
		Clients: map[*Client]bool{
			client:           true,
			developerToReset: true,
		},
		broadcast: make(chan *OutgoingWebsocketMessage),
		destroy:   nil,
	}
	go room.Run()

	msg := newOutgoingWebsocketMessage(newRound, nil)
	room.broadcast <- msg

	// four messages due to two clients à 2 messages
	gotFirstClientMessage := <-clientSendChannel
	gotSecondClientMessage := <-clientSendChannel
	gotThirdClientMessage := <-clientSendChannel
	gotFourthClientMessage := <-clientSendChannel

	assert.DeepEqual(t, gotFirstClientMessage, msg)
	assert.DeepEqual(t, gotSecondClientMessage, newUsers(room.Clients))
	assert.DeepEqual(t, gotThirdClientMessage, msg)
	assert.DeepEqual(t, gotFourthClientMessage, newUsers(room.Clients))

	assert.False(t, room.IsInProgress())
	assert.Equal(t, developerToReset.Guess(), 0)
}

func TestRoom_Run_BroadcastLeaveWhenRoomInProgress(t *testing.T) {
	clientSendChannel := make(chan *OutgoingWebsocketMessage)
	broadcastChannel := make(chan *OutgoingWebsocketMessage)
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
		Id:         uuid.New(),
		inProgress: true,
		leave:      nil,
		join:       nil,
		Clients: map[*Client]bool{
			client:           true,
			developerToReset: true,
		},
		broadcast: broadcastChannel,
		destroy:   nil,
	}
	go room.Run()

	msg := newOutgoingWebsocketMessage(leave, client.Name)
	room.broadcast <- msg
	gotClientMsg := <-clientSendChannel
	<-clientSendChannel

	assert.DeepEqual(t, gotClientMsg, newOutgoingWebsocketMessage(newRound, nil))
	assert.False(t, room.inProgress)
	assert.Equal(t, developerToReset.Guess(), 0)
}

func TestRoom_lock(t *testing.T) {
	key := uuid.New()
	room := &Room{
		Id:             uuid.New(),
		inProgress:     true,
		leave:          nil,
		join:           nil,
		Clients:        make(map[*Client]bool),
		broadcast:      make(chan *OutgoingWebsocketMessage),
		destroy:        nil,
		NameOfCreator:  "Bla",
		key:            key,
		HashedPassword: make([]byte, 0),
	}

	got := room.lock("Bla", "top secret", key.String())

	assert.True(t, got)
	assert.True(t, room.IsLocked())
	assert.False(t, len(room.HashedPassword) == 0)
}

func TestRoom_lock_WhenLockingFails(t *testing.T) {
	key := uuid.New()
	room := &Room{
		Id:             uuid.New(),
		inProgress:     true,
		leave:          nil,
		join:           nil,
		Clients:        make(map[*Client]bool),
		broadcast:      make(chan *OutgoingWebsocketMessage),
		destroy:        nil,
		NameOfCreator:  "Bla",
		key:            key,
		HashedPassword: make([]byte, 0),
	}

	got := room.lock("ABC", "top secret", key.String())

	assert.False(t, got)
}

func TestRoom_open_WhenUserNotCreator(t *testing.T) {
	id := uuid.New()
	room := &Room{
		Id:            uuid.New(),
		inProgress:    false,
		NameOfCreator: "some user",
		key:           id,
	}
	got := room.open("invalid user", id.String())

	assert.False(t, got)
}

func TestRoom_open_WhenKeyIsWrong(t *testing.T) {
	room := &Room{
		Id:            uuid.New(),
		inProgress:    false,
		NameOfCreator: "some user",
		key:           uuid.New(),
	}

	got := room.open("some user", "incorrect Key")

	assert.False(t, got)
}

func TestRoom_lock_WhenLockingFailsDueToHashingFails(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	key := uuid.New()
	room := &Room{
		Id:             uuid.New(),
		inProgress:     true,
		leave:          nil,
		join:           nil,
		Clients:        make(map[*Client]bool),
		broadcast:      make(chan *OutgoingWebsocketMessage),
		destroy:        nil,
		NameOfCreator:  "Bla",
		key:            key,
		HashedPassword: make([]byte, 0),
		logger:         logger,
	}

	got := room.lock("ABC", strings.Repeat("bla", 90), key.String())
	wantedLog := "failed to hash password"

	assert.False(t, got)
	assert.StringContains(t, logBuffer.String(), wantedLog)
}

func TestRoom_AsOverview(t *testing.T) {
	created := time.Now()
	tests := []struct {
		name    string
		id      uuid.UUID
		clients map[*Client]bool
		created time.Time
		want    Overview
	}{
		{
			name: "correct representation",
			id:   uuid.MustParse("cdf64eb1-48d5-4ef5-b0c5-887b893c85dd"),
			clients: map[*Client]bool{
				&Client{}: true,
			},
			created: created,
			want: Overview{
				Id:          uuid.MustParse("cdf64eb1-48d5-4ef5-b0c5-887b893c85dd"),
				PlayerCount: 1,
				Created:     created,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := &Room{
				Id:      tt.id,
				Clients: tt.clients,
				Created: created,
			}
			assert.DeepEqual(t, room.AsOverview(), tt.want)
		})
	}
}

func TestRoom_State(t *testing.T) {
	tests := []struct {
		name            string
		inProgress      bool
		hashedPassword  []byte
		possibleGuesses []GuessConfigEntry
		issues          []*Issue
		want            State
	}{
		{
			name:           "correct representation",
			inProgress:     false,
			hashedPassword: make([]byte, 0),
			issues:         make([]*Issue, 0),
			possibleGuesses: []GuessConfigEntry{
				{
					Guess:       0,
					Description: "A",
				},
				{
					Guess:       1,
					Description: "B",
				},
			},
			want: State{
				InProgress: false,
				IsLocked:   false,
				Issues:     make([]*Issue, 0),
				PossibleGuesses: []GuessConfigEntry{
					{
						Guess:       0,
						Description: "A",
					},
					{
						Guess:       1,
						Description: "B",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := &Room{
				inProgress:     tt.inProgress,
				HashedPassword: tt.hashedPassword,
				issues:         tt.issues,
				GuessConfig: &GuessConfig{
					Guesses: tt.possibleGuesses,
				},
			}
			assert.DeepEqual(t, room.State(), tt.want)
		})
	}
}

func TestRoom_addIssue(t *testing.T) {
	tests := []struct {
		name       string
		issueToAdd string
		want       []*Issue
	}{
		{
			name:       "default guess for new issue",
			issueToAdd: "Hello World",
			want: []*Issue{
				{
					Title: "Hello World",
					Guess: -1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := &Room{
				issues: make([]*Issue, 0),
			}

			room.addIssue(tt.issueToAdd)

			assert.DeepEqual(t, room.issues, tt.want)
		})
	}
}
