package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/Hydoc/estimation-poker/backend/internal"
	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func TestApplication_createNewRoom(t *testing.T) {
	tests := []struct {
		name               string
		body               map[string]any
		expectedStatusCode int
	}{
		{
			name: "create new room",
			body: map[string]any{
				"creator": "Tester",
				"guesses": map[int]string{
					1: "up to 4 hours",
					2: "up to 1 day",
				},
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name: "not create for invalid data",
			body: map[string]any{
				"creator": 2,
			},
			expectedStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t, make(map[uuid.UUID]*internal.Room))

			ts := newTestServer(t, app.routes())
			defer ts.Close()

			response := ts.postJSON(t, "/v1/room", tt.body)
			gotContentType := response.headers.Get("Content-Type")

			assert.Equal(t, response.status, tt.expectedStatusCode)
			assert.Equal(t, gotContentType, "application/json")
		})
	}
}

func TestApplication_handleFetchRoomMetadata(t *testing.T) {
	tests := []struct {
		name       string
		roomId     string
		rooms      map[uuid.UUID]*internal.Room
		wantStatus int
		wantBody   envelope
	}{
		{
			name:   "room exists and is not locked",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					HashedPassword: make([]byte, 0),
					Issues:         make([]internal.Issue, 0),
					GuessConfig:    &internal.GuessConfig{},
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   envelope{"exists": true, "isLocked": false},
		},
		{
			name:       "does not exist",
			roomId:     "bd284176-7a5d-4443-b0e0-5058c3e07853",
			rooms:      make(map[uuid.UUID]*internal.Room),
			wantStatus: http.StatusOK,
			wantBody:   envelope{"exists": false, "isLocked": false},
		},
		{
			name:       "room id is invalid",
			roomId:     "invalid-uuid",
			rooms:      make(map[uuid.UUID]*internal.Room),
			wantStatus: http.StatusBadRequest,
			wantBody:   envelope{"error": "invalid id parameter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t, tt.rooms)
			ts := newTestServer(t, app.routes())
			defer ts.Close()

			response := ts.get(t, fmt.Sprintf("/v1/room/%s/metadata", tt.roomId))

			var got envelope
			json.Unmarshal(response.body, &got)

			assert.DeepEqual(t, got, tt.wantBody)
		})
	}
}

func TestApplication_handleFetchConnectionState(t *testing.T) {
	tests := []struct {
		name           string
		roomId         string
		rooms          map[uuid.UUID]*internal.Room
		body           map[string]any
		wantStatusCode int
		wantState      internal.ConnectionState
	}{
		{
			name:   "get connection state for fresh room",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					HashedPassword: make([]byte, 0),
					Clients:        make(map[*internal.Client]bool),
				},
			},
			body: map[string]any{
				"username": "test",
				"password": "",
			},
			wantStatusCode: http.StatusOK,
			wantState: internal.ConnectionState{
				CanConnect: true,
				Reason:     "",
			},
		},
		{

			name:   "get connection state when password matches",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					HashedPassword: []byte("$2a$12$SUOHh5BqzhO5nDMxLhvv6.SZiq/6wM5A8zYg285ZFtkF5meuG17Nm"),
					Clients:        make(map[*internal.Client]bool),
				},
			},
			body: map[string]any{
				"username": "test",
				"password": "Hello",
			},
			wantStatusCode: http.StatusOK,
			wantState: internal.ConnectionState{
				CanConnect: true,
				Reason:     "",
			},
		},
		//{
		//	name:   "room in progress",
		//	roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
		//	rooms: map[uuid.UUID]*internal.Room{
		//		uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): CreateRoomInProgress(),
		//	},
		//	body: map[string]any{
		//		"username": "test",
		//		"password": "",
		//	},
		//	wantStatusCode: http.StatusOK,
		//	wantState: internal.ConnectionState{
		//		CanConnect: false,
		//		Reason:     internal.ErrRoundStarted.Error(),
		//	},
		//},
		{
			name:   "room is locked and not password provided",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					HashedPassword: make([]byte, 1),
				},
			},
			body: map[string]any{
				"username": "test",
				"password": "",
			},
			wantStatusCode: http.StatusOK,
			wantState: internal.ConnectionState{
				CanConnect: false,
				Reason:     internal.ErrWrongPassword.Error(),
			},
		},
		{
			name:   "room is locked and password does not match",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					HashedPassword: []byte("$2a$12$SUOHh5BqzhO5nDMxLhvv6.SZiq/6wM5A8zYg285ZFtkF5meuG17Nm"),
				},
			},
			body: map[string]any{
				"username": "test",
				"password": "asd",
			},
			wantStatusCode: http.StatusOK,
			wantState: internal.ConnectionState{
				CanConnect: false,
				Reason:     internal.ErrWrongPassword.Error(),
			},
		},
		{
			name:   "username already taken",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					HashedPassword: make([]byte, 0),
					Clients: map[*internal.Client]bool{
						&internal.Client{
							Name: "test",
						}: true,
					},
				},
			},
			body: map[string]any{
				"username": "test",
				"password": "asd",
			},
			wantStatusCode: http.StatusOK,
			wantState: internal.ConnectionState{
				CanConnect: false,
				Reason:     internal.ErrUsernameTaken.Error(),
			},
		},
		{
			name:           "invalid room id",
			roomId:         "invalid",
			rooms:          make(map[uuid.UUID]*internal.Room),
			body:           nil,
			wantStatusCode: http.StatusBadRequest,
			wantState:      internal.ConnectionState{},
		},
		{
			name:           "not found room",
			roomId:         "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms:          make(map[uuid.UUID]*internal.Room),
			body:           nil,
			wantStatusCode: http.StatusNotFound,
			wantState:      internal.ConnectionState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t, tt.rooms)
			ts := newTestServer(t, app.routes())
			defer ts.Close()

			response := ts.postJSON(t, fmt.Sprintf("/v1/room/%s/connection-state", tt.roomId), tt.body)

			assert.Equal(t, response.headers.Get("Content-Type"), "application/json")
			assert.Equal(t, response.status, tt.wantStatusCode)

			var got internal.ConnectionState
			json.Unmarshal(response.body, &got)
			assert.DeepEqual(t, got, tt.wantState)
		})
	}
}

func TestApplication_handleFetchRoomState(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectation    internal.State
		rooms          map[uuid.UUID]*internal.Room
		room           string
	}{
		{
			name:           "not in progress when rooms are empty",
			expectedStatus: http.StatusNotFound,
			expectation:    internal.State{},
			rooms:          map[uuid.UUID]*internal.Room{},
			room:           "9c874aaa-c628-4688-a72d-0b1afc708a7d",
		},
		//{
		//	name:           "in progress when room is set",
		//	expectedStatus: http.StatusOK,
		//	expectation: internal.State{
		//		InProgress:      true,
		//		IsLocked:        false,
		//		Issues:          make([]internal.Issue, 0),
		//		PossibleGuesses: nil,
		//	},
		//	rooms: map[uuid.UUID]*internal.Room{
		//		uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
		//			inProgress:     true,
		//			HashedPassword: make([]byte, 0),
		//			Issues:         make([]internal.Issue, 0),
		//			GuessConfig:    &internal.GuessConfig{},
		//		},
		//	},
		//	room: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t, tt.rooms)
			ts := newTestServer(t, app.routes())
			defer ts.Close()

			response := ts.get(t, fmt.Sprintf("/v1/room/%s/state", tt.room))

			var got internal.State
			json.Unmarshal(response.body, &got)

			gotContentType := response.headers.Get("Content-Type")

			assert.Equal(t, response.status, tt.expectedStatus)
			assert.Equal(t, gotContentType, "application/json")
			assert.DeepEqual(t, got, tt.expectation)
		})
	}
}

func TestApplication_handleFetchActiveRooms(t *testing.T) {
	tests := []struct {
		name  string
		rooms func() map[uuid.UUID]*internal.Room
		want  map[string][]map[string]any
	}{
		{
			name: "with multiple active rooms",
			rooms: func() map[uuid.UUID]*internal.Room {
				firstDate, err := time.Parse("2006-01-02", "2026-01-01")
				if err != nil {
					t.Fatal(err)
				}

				secondDate, err := time.Parse("2006-01-02", "2026-02-01")
				if err != nil {
					t.Fatal(err)
				}
				return map[uuid.UUID]*internal.Room{
					uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
						Id:      uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"),
						Created: secondDate,
					},
					uuid.MustParse("50e15380-1475-4ec6-abb0-f1e22929a8e5"): {
						Id:      uuid.MustParse("50e15380-1475-4ec6-abb0-f1e22929a8e5"),
						Created: firstDate,
					},
				}
			},
			want: map[string][]map[string]any{
				"rooms": {
					{
						"id":          "50e15380-1475-4ec6-abb0-f1e22929a8e5",
						"playerCount": float64(0),
					},
					{
						"id":          "9c874aaa-c628-4688-a72d-0b1afc708a7d",
						"playerCount": float64(0),
					},
				},
			},
		},
		{
			name: "multiple rooms but one is locked",
			rooms: func() map[uuid.UUID]*internal.Room {
				return map[uuid.UUID]*internal.Room{
					uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
						Id:             uuid.MustParse("9c874aaa-c628-4688-a72d-0b1afc708a7d"),
						HashedPassword: []byte("does not matter"),
					},
					uuid.MustParse("50e15380-1475-4ec6-abb0-f1e22929a8e5"): {
						Id: uuid.MustParse("50e15380-1475-4ec6-abb0-f1e22929a8e5"),
					},
				}
			},
			want: map[string][]map[string]any{
				"rooms": {
					{
						"id":          "50e15380-1475-4ec6-abb0-f1e22929a8e5",
						"playerCount": float64(0),
					},
				},
			},
		},
		{
			name: "no rooms",
			rooms: func() map[uuid.UUID]*internal.Room {
				return make(map[uuid.UUID]*internal.Room)
			},
			want: map[string][]map[string]any{
				"rooms": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t, tt.rooms())

			ts := newTestServer(t, app.routes())
			defer ts.Close()

			response := ts.get(t, "/v1/rooms")

			var got map[string][]map[string]any
			json.Unmarshal(response.body, &got)

			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestApplication_handleWs(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		rooms          map[uuid.UUID]*internal.Room
		expectedError  map[string]string
		expectedRoomId string
		expectedRole   string
		expectedStatus int
	}{
		{
			name: "connect as developer",
			url:  "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/developer?name=Test",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"): {
					Id: uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"),
				},
			},
			expectedError:  nil,
			expectedStatus: 101,
			expectedRoomId: "ffb25a3d-a5db-42b7-9733-345f61167077",
			expectedRole:   internal.Developer,
		},
		{
			name: "connect as product owner",
			url:  "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=Test",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"): {
					Id: uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"),
				},
			},
			expectedError:  nil,
			expectedStatus: 101,
			expectedRoomId: "ffb25a3d-a5db-42b7-9733-345f61167077",
			expectedRole:   internal.ProductOwner,
		},
		{
			name:  "not connecting due to name too long",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=whateverthisisitiswaytoooooooooooolong",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"error": "name must be smaller or equal to 15",
			},
			expectedStatus: 400,
		},
		{
			name:  "not connecting due to missing name",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"message": "name is missing in query",
			},
			expectedStatus: 400,
			expectedRoomId: "ffb25a3d-a5db-42b7-9733-345f61167077",
			expectedRole:   internal.ProductOwner,
		},
		{
			name:  "not connecting due to invalid roomId not found",
			url:   "/v1/room/invalid/product-owner?name=test",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"error": "invalid id parameter",
			},
			expectedStatus: 400,
		},
		{
			name:  "not connecting because room not found",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=test",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"error": "the requested resource could not be found",
			},
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t, tt.rooms)

			ts := newTestServer(t, app.routes())
			defer ts.Close()

			url := "ws" + strings.TrimPrefix(ts.URL, "http") + tt.url
			_, response, _ := websocket.Dial(context.Background(), url, nil)

			assert.Equal(t, response.StatusCode, tt.expectedStatus)

			if tt.expectedError != nil {
				var got map[string]string
				json.NewDecoder(response.Body).Decode(&got)
				assert.DeepEqual(t, got, tt.expectedError)
				return
			}
		})
	}
}

func TestApplication_ListenForRoomDestroy(t *testing.T) {
	destroyChannel := make(chan uuid.UUID)
	roomToDestroy := uuid.MustParse("e8563735-ca82-4fad-b9fc-4942c5b0cdb0")
	app := &application{
		guessConfig: &internal.GuessConfig{},
		rooms: map[uuid.UUID]*internal.Room{
			roomToDestroy: {
				Id: roomToDestroy,
			},
		},
		destroyRoom: destroyChannel,
	}
	go app.listenForRoomDestroy(context.Background())

	app.destroyRoom <- roomToDestroy

	app.mu.Lock()
	defer app.mu.Unlock()
	if _, ok := app.rooms[roomToDestroy]; ok {
		t.Error("expected app to not have room")
	}
}
