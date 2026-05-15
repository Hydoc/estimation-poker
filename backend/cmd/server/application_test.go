package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hydoc/go-message"
	"github.com/coder/websocket"

	"github.com/Hydoc/estimation-poker/backend/internal"
	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func TestApplication_handleFetchRoomState(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectation    internal.State
		rooms          map[internal.RoomId]*internal.Room
		room           string
	}{
		{
			name:           "not in progress when rooms are empty",
			expectedStatus: http.StatusNotFound,
			expectation:    internal.State{},
			rooms:          map[internal.RoomId]*internal.Room{},
			room:           "9c874aaa-c628-4688-a72d-0b1afc708a7d",
		},
		{
			name:           "in progress when room is set",
			expectedStatus: http.StatusOK,
			expectation: internal.State{
				InProgress:      true,
				IsLocked:        false,
				Issues:          make([]internal.Issue, 0),
				PossibleGuesses: nil,
			},
			rooms: map[internal.RoomId]*internal.Room{
				"9c874aaa-c628-4688-a72d-0b1afc708a7d": {
					InProgress:     true,
					HashedPassword: make([]byte, 0),
					Issues:         make([]internal.Issue, 0),
					GuessConfig:    &internal.GuessConfig{},
				},
			},
			room: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
		},
	}

	for _, test := range tests {
		app := &application{
			rooms: test.rooms,
		}

		t.Run(test.name, func(t *testing.T) {
			router := app.routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/room/%s/state", test.room), nil)

			router.ServeHTTP(recorder, request)

			var got internal.State
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")

			assert.Equal(t, recorder.Code, test.expectedStatus)
			assert.Equal(t, gotContentType, "application/json")
			assert.DeepEqual(t, got, test.expectation)
		})
	}
}

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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &application{
				rooms: make(map[internal.RoomId]*internal.Room),
			}

			body, err := json.Marshal(test.body)

			if err != nil {
				t.Fatal(err)
			}

			router := app.routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/v1/room", bytes.NewReader(body))

			router.ServeHTTP(recorder, request)

			gotContentType := recorder.Header().Get("Content-Type")

			assert.Equal(t, recorder.Code, test.expectedStatusCode)
			assert.Equal(t, gotContentType, "application/json")
		})
	}
}

func TestApplication_handleFetchUsers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	room := internal.NewRoom("9c874aaa-c628-4688-a72d-0b1afc708a7d", make(chan<- internal.RoomId), "", logger, new(internal.GuessConfig))
	bus := message.NewBus()
	dev := internal.NewClient("B", internal.Developer, room, nil, bus, logger)
	otherDev := internal.NewClient("Another", internal.Developer, room, nil, bus, logger)
	devWithEqualLetter := internal.NewClient("Also a dev", internal.Developer, room, nil, bus, logger)
	productOwner := internal.NewClient("Another one", internal.ProductOwner, room, nil, bus, logger)
	otherProductOwner := internal.NewClient("Also a po", internal.ProductOwner, room, nil, bus, logger)

	tests := []struct {
		name        string
		roomId      string
		rooms       map[internal.RoomId]*internal.Room
		expectation []map[string]any
	}{
		{
			name:   "some users in the same room",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					Id:         "9c874aaa-c628-4688-a72d-0b1afc708a7d",
					InProgress: false,
					Clients: map[*internal.Client]bool{
						dev:                true,
						otherDev:           true,
						devWithEqualLetter: true,
						productOwner:       true,
						otherProductOwner:  true,
					},
				},
			},
			expectation: []map[string]any{
				{
					"name":   "Also a dev",
					"isDone": false,
					"role":   internal.Developer,
				},
				{
					"name": "Also a po",
					"role": internal.ProductOwner,
				},
				{
					"name":   "Another",
					"isDone": false,
					"role":   internal.Developer,
				},
				{
					"name": "Another one",
					"role": internal.ProductOwner,
				},
				{
					"name":   "B",
					"isDone": false,
					"role":   internal.Developer,
				},
			},
		},
		{
			name:        "no clients",
			roomId:      "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms:       make(map[internal.RoomId]*internal.Room),
			expectation: []map[string]any{},
		},
		{
			name:   "one dev client",
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					Clients: map[*internal.Client]bool{
						dev: true,
					},
				},
			},
			expectation: []map[string]any{
				{
					"name":   "B",
					"isDone": false,
					"role":   internal.Developer,
				},
			},
		},
		{
			name: "one po client",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					Clients: map[*internal.Client]bool{
						productOwner: true,
					},
				},
			},
			roomId: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
			expectation: []map[string]any{
				{
					"name": "Another one",
					"role": internal.ProductOwner,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &application{
				guessConfig: &internal.GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: nil,
			}
			router := app.routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/room/%s/users", test.roomId), nil)

			router.ServeHTTP(recorder, request)

			var got []map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")

			assert.Equal(t, gotContentType, "application/json")
			assert.DeepEqual(t, got, test.expectation)
		})
	}
}

func TestApplication_handleWs(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		rooms          map[internal.RoomId]*internal.Room
		expectedError  map[string]string
		expectedRoomId string
		expectedRole   string
		expectedStatus int
	}{
		{
			name: "connect as developer",
			url:  "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/developer?name=Test",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("ffb25a3d-a5db-42b7-9733-345f61167077"): {
					Id:         "ffb25a3d-a5db-42b7-9733-345f61167077",
					InProgress: false,
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
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("ffb25a3d-a5db-42b7-9733-345f61167077"): {
					Id:         "ffb25a3d-a5db-42b7-9733-345f61167077",
					InProgress: false,
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
			rooms: make(map[internal.RoomId]*internal.Room),
			expectedError: map[string]string{
				"error": "name must be smaller or equal to 15",
			},
			expectedStatus: 400,
		},
		{
			name:  "not connecting due to missing name",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=",
			rooms: make(map[internal.RoomId]*internal.Room),
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
			rooms: make(map[internal.RoomId]*internal.Room),
			expectedError: map[string]string{
				"error": "invalid id parameter",
			},
			expectedStatus: 400,
		},
		{
			name:  "not connecting because room not found",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=test",
			rooms: make(map[internal.RoomId]*internal.Room),
			expectedError: map[string]string{
				"error": "the requested resource could not be found",
			},
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &application{
				guessConfig: &internal.GuessConfig{},
				rooms:       tt.rooms,
				destroyRoom: make(chan internal.RoomId),
			}
			router := app.routes()

			server := httptest.NewServer(router)
			defer server.Close()

			url := "ws" + strings.TrimPrefix(server.URL, "http") + tt.url
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

func TestApplication_handleFetchActiveRooms(t *testing.T) {
	tests := []struct {
		name  string
		rooms map[internal.RoomId]*internal.Room
		want  map[string][]map[string]any
	}{
		{
			name: "with multiple active rooms",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("50e15380-1475-4ec6-abb0-f1e22929a8e5"): {
					Id: "50e15380-1475-4ec6-abb0-f1e22929a8e5",
				},
				internal.RoomId("9c874aaa-c628-4688-a72d-0b1afc708a7d"): {
					Id: "9c874aaa-c628-4688-a72d-0b1afc708a7d",
				},
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
			name:  "no rooms",
			rooms: make(map[internal.RoomId]*internal.Room),
			want: map[string][]map[string]any{
				"rooms": {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &application{
				guessConfig: &internal.GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: nil,
			}
			router := app.routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/v1/rooms", nil)

			router.ServeHTTP(recorder, request)

			var got map[string][]map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &got)

			assert.DeepEqual(t, got, test.want)
		})
	}
}

func TestApplication_ListenForRoomDestroy(t *testing.T) {
	destroyChannel := make(chan internal.RoomId)
	roomToDestroy := internal.RoomId("Test")
	app := &application{
		guessConfig: &internal.GuessConfig{},
		rooms: map[internal.RoomId]*internal.Room{
			roomToDestroy: {
				Id:         roomToDestroy,
				InProgress: false,
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
