package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Hydoc/guess-dev/backend/internal"
	"github.com/Hydoc/guess-dev/backend/internal/assert"
)

func TestApplication_handleRoundInRoomInProgress(t *testing.T) {
	tests := []struct {
		name        string
		expectation map[string]bool
		rooms       map[internal.RoomId]*internal.Room
		room        string
	}{
		{
			name: "not in progress when rooms are empty",
			expectation: map[string]bool{
				"inProgress": false,
				"isLocked":   false,
			},
			rooms: map[internal.RoomId]*internal.Room{},
			room:  "1",
		},
		{
			name: "in progress when room is set",
			expectation: map[string]bool{
				"inProgress": true,
				"isLocked":   false,
			},
			rooms: map[internal.RoomId]*internal.Room{
				"1": {
					InProgress: true,
				},
			},
			room: "1",
		},
	}

	for _, test := range tests {
		app := &application{
			rooms: test.rooms,
		}

		t.Run(test.name, func(t *testing.T) {
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/estimation/room/%s/state", test.room), nil)

			router.ServeHTTP(recorder, request)

			var got map[string]bool
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")

			assert.Equal(t, gotContentType, "application/json")
			assert.DeepEqual(t, got, test.expectation)
		})
	}
}

func TestApplication_handleUserInRoomExists(t *testing.T) {
	client := &internal.Client{
		Name: "Test",
	}
	roomId := "1"
	tests := []struct {
		url               string
		name              string
		rooms             map[internal.RoomId]*internal.Room
		expectation       map[string]interface{}
		expectedErrorBody map[string]interface{}
		statusCode        int
	}{
		{
			name: "not find client when clients are empty",
			url:  fmt.Sprintf("/api/estimation/room/%s/users/exists?name=Bla", roomId),
			expectation: map[string]interface{}{
				"exists": false,
			},
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId(roomId): {
					Clients: make(map[*internal.Client]bool),
				},
			},
			statusCode: 200,
		},
		{
			name: "find client when client exists",
			url:  fmt.Sprintf("/api/estimation/room/%s/users/exists?name=%s", roomId, client.Name),
			expectation: map[string]interface{}{
				"exists": true,
			},
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId(roomId): {
					Clients: map[*internal.Client]bool{
						client: true,
					},
				},
			},
			statusCode: 409,
		},
		{
			name:        "error when trying to access without name",
			url:         fmt.Sprintf("/api/estimation/room/%s/users/exists?name=", roomId),
			expectation: nil,
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId(roomId): {
					Clients: map[*internal.Client]bool{
						client: true,
					},
				},
			},
			expectedErrorBody: map[string]interface{}{
				"message": "name is missing in query",
			},
			statusCode: 400,
		},
		{
			name: "not find client when no room with id was found",
			url:  fmt.Sprintf("/api/estimation/room/%s/users/exists?name=Test", roomId),
			expectation: map[string]interface{}{
				"exists": false,
			},
			rooms:             map[internal.RoomId]*internal.Room{},
			expectedErrorBody: nil,
			statusCode:        200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &application{
				rooms: test.rooms,
			}
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.url, nil)

			router.ServeHTTP(recorder, request)

			var got map[string]interface{}
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")

			assert.Equal(t, test.statusCode, recorder.Code)
			assert.Equal(t, gotContentType, "application/json")
			if test.expectedErrorBody != nil {
				assert.DeepEqual(t, got, test.expectedErrorBody)
			}

			if test.expectation != nil {
				assert.DeepEqual(t, got, test.expectation)
			}
		})
	}
}

func TestApplication_handleFetchUsers(t *testing.T) {
	room := internal.NewRoom("1", make(chan<- internal.RoomId), "")
	dev := internal.NewClient("B", internal.Developer, room, nil)
	otherDev := internal.NewClient("Another", internal.Developer, room, nil)
	devWithEqualLetter := internal.NewClient("Also a dev", internal.Developer, room, nil)
	productOwner := internal.NewClient("Another one", internal.ProductOwner, room, nil)
	otherProductOwner := internal.NewClient("Also a po", internal.ProductOwner, room, nil)

	tests := []struct {
		name        string
		roomId      string
		rooms       map[internal.RoomId]*internal.Room
		expectation []map[string]any
	}{
		{
			name:   "some users in the same room",
			roomId: "1",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("1"): {
					Id:         "1",
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
			roomId:      "1",
			rooms:       make(map[internal.RoomId]*internal.Room),
			expectation: []map[string]any{},
		},
		{
			name:   "one dev client",
			roomId: "1",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("1"): {
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
				internal.RoomId("1"): {
					Clients: map[*internal.Client]bool{
						productOwner: true,
					},
				},
			},
			roomId: "1",
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
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/estimation/room/%s/users", test.roomId), nil)

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
			name: "register as developer",
			url:  "/api/estimation/room/1/developer?name=Test",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("1"): {
					Id:         "1",
					InProgress: false,
					Join:       make(chan *internal.Client),
					Broadcast:  make(chan *internal.Message),
				},
			},
			expectedError:  nil,
			expectedStatus: -1,
			expectedRoomId: "1",
			expectedRole:   internal.Developer,
		},
		{
			name: "register as product owner",
			url:  "/api/estimation/room/1/product-owner?name=Test",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("1"): {
					Id:         "1",
					InProgress: false,
					Join:       make(chan *internal.Client),
					Broadcast:  make(chan *internal.Message),
				},
			},
			expectedError:  nil,
			expectedStatus: -1,
			expectedRoomId: "1",
			expectedRole:   internal.ProductOwner,
		},
		{
			name:  "not registering due to missing name",
			url:   "/api/estimation/room/1/product-owner?name=",
			rooms: make(map[internal.RoomId]*internal.Room),
			expectedError: map[string]string{
				"message": "name is missing in query",
			},
			expectedStatus: 400,
			expectedRoomId: "1",
			expectedRole:   internal.ProductOwner,
		},
		{
			name:  "not registering due to name too long",
			url:   "/api/estimation/room/1/product-owner?name=mynameiswaytoolongitshouldnotbecreated",
			rooms: make(map[internal.RoomId]*internal.Room),
			expectedError: map[string]string{
				"message": "name and room must be smaller or equal to 15",
			},
			expectedStatus: 400,
			expectedRoomId: "1",
			expectedRole:   internal.ProductOwner,
		},
		{
			name:  "not registering due to roomId too long",
			url:   "/api/estimation/room/whateverthatroomisitiswaytoolong/product-owner?name=nameok",
			rooms: make(map[internal.RoomId]*internal.Room),
			expectedError: map[string]string{
				"message": "name and room must be smaller or equal to 15",
			},
			expectedStatus: 400,
			expectedRoomId: "1",
			expectedRole:   internal.ProductOwner,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedMsg := internal.NewJoin()
			app := &application{
				guessConfig: &internal.GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: make(chan internal.RoomId),
			}
			router := app.Routes()

			server := httptest.NewServer(router)
			defer server.Close()

			url := "ws" + strings.TrimPrefix(server.URL, "http") + test.url
			_, response, _ := websocket.Dial(context.Background(), url, nil)

			if test.expectedError != nil {
				var got map[string]string
				json.NewDecoder(response.Body).Decode(&got)
				assert.DeepEqual(t, got, test.expectedError)
				return
			}

			registeredClient := <-app.rooms[internal.RoomId(test.expectedRoomId)].Join
			broadcastedMsg := <-app.rooms[internal.RoomId(test.expectedRoomId)].Broadcast

			assert.DeepEqual(t, broadcastedMsg, expectedMsg)
			assert.Equal(t, registeredClient.Role, test.expectedRole)
		})
	}
}

func TestApplication_handleWs_CreatingNewRoom(t *testing.T) {
	app := &application{
		rooms:       make(map[internal.RoomId]*internal.Room),
		destroyRoom: make(chan internal.RoomId),
	}
	roomId := "Test"
	expectedRoom := internal.NewRoom(internal.RoomId(roomId), app.destroyRoom, "")
	router := app.Routes()

	server := httptest.NewServer(router)
	defer server.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		url := "ws" + strings.TrimPrefix(server.URL, "http") + fmt.Sprintf("/api/estimation/room/%s/developer?name=Test", roomId)
		websocket.Dial(context.Background(), url, nil)
	}()

	wg.Wait()

	app.roomMu.Lock()
	got := app.rooms[internal.RoomId(roomId)]
	app.roomMu.Unlock()

	assert.Equal(t, got.Id, expectedRoom.Id)
}

func TestApplication_handleFetchActiveRooms(t *testing.T) {
	tests := []struct {
		name  string
		rooms map[internal.RoomId]*internal.Room
		want  []string
	}{
		{
			name: "with multiple active rooms",
			rooms: map[internal.RoomId]*internal.Room{
				internal.RoomId("Blub"): {
					Id: "Blub",
				},
				internal.RoomId("Test"): {
					Id: "Test",
				},
			},
			want: []string{"Blub", "Test"},
		},
		{
			name:  "no rooms",
			rooms: make(map[internal.RoomId]*internal.Room),
			want:  []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &application{
				guessConfig: &internal.GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: nil,
			}
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/estimation/room/rooms", nil)

			router.ServeHTTP(recorder, request)

			var got []string
			json.Unmarshal(recorder.Body.Bytes(), &got)

			assert.DeepEqual(t, got, test.want)
		})
	}
}

func TestApplication_handlePossibleGuesses(t *testing.T) {
	tests := []struct {
		name   string
		config *internal.GuessConfig
		want   []map[string]interface{}
	}{
		{
			name: "multiple guesses",
			config: &internal.GuessConfig{
				Guesses: []internal.GuessConfigEntry{
					{
						Guess:       1,
						Description: "Test 1",
					},
					{
						Guess:       3,
						Description: "Test 3",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"guess":       float64(1),
					"description": "Test 1",
				},
				{
					"guess":       float64(3),
					"description": "Test 3",
				},
			},
		},
		{
			name: "one guess",
			config: &internal.GuessConfig{
				Guesses: []internal.GuessConfigEntry{
					{
						Guess:       1,
						Description: "Test 1",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"guess":       float64(1),
					"description": "Test 1",
				},
			},
		},
		{
			name: "no guess",
			config: &internal.GuessConfig{
				Guesses: make([]internal.GuessConfigEntry, 0),
			},
			want: make([]map[string]interface{}, 0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &application{
				guessConfig: test.config,
			}
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/estimation/possible-guesses", nil)

			router.ServeHTTP(recorder, request)

			var got []map[string]interface{}
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

	app.roomMu.Lock()
	defer app.roomMu.Unlock()
	if _, ok := app.rooms[roomToDestroy]; ok {
		t.Error("expected app to not have room")
	}
}

func TestApplication_handleRoomAuthenticate(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
		wantBody map[string]any
		request  string
		body     func() *bytes.Buffer
	}{
		{
			name:     "forbidden when room not found",
			wantCode: http.StatusForbidden,
			wantBody: nil,
			body: func() *bytes.Buffer {
				return nil
			},
			request: "/api/estimation/room/12/authenticate",
		},
		{
			name:     "ok = false when body can not be decoded",
			wantCode: http.StatusOK,
			wantBody: envelope{"ok": false},
			body: func() *bytes.Buffer {
				return bytes.NewBuffer([]byte(""))
			},
			request: "/api/estimation/room/1/authenticate",
		},
		{
			name:     "ok = true when password matches",
			wantCode: http.StatusOK,
			wantBody: envelope{"ok": true},
			body: func() *bytes.Buffer {
				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(map[string]string{
					"password": "helo world",
				})
				if err != nil {
					t.Error(err)
				}
				return &buf
			},
			request: "/api/estimation/room/1/authenticate",
		},
		{
			name:     "ok = false when password does not match",
			wantCode: http.StatusOK,
			wantBody: envelope{"ok": false},
			body: func() *bytes.Buffer {
				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(map[string]string{
					"password": "invalid",
				})
				if err != nil {
					t.Error(err)
				}
				return &buf
			},
			request: "/api/estimation/room/1/authenticate",
		},
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("helo world"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("error hashing password: %v", err)
	}
	app := &application{
		guessConfig: &internal.GuessConfig{},
		rooms: map[internal.RoomId]*internal.Room{
			"1": {
				Id:             "1",
				HashedPassword: hashedPassword,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, test.request, nil)
			body := test.body()

			if body != nil {
				request = httptest.NewRequest(http.MethodPost, test.request, body)
			}

			router.ServeHTTP(recorder, request)

			var got map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotCode := recorder.Code

			assert.Equal(t, gotCode, test.wantCode)
			assert.DeepEqual(t, got, test.wantBody)
		})
	}
}

func TestApplication_handleFetchPermissions(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name     string
		wantCode int
		wantBody map[string]map[string]map[string]any
		request  string
	}{
		{
			name:     "not found when room not found",
			wantCode: http.StatusNotFound,
			wantBody: nil,
			request:  "/api/estimation/room/1244132/test/permissions",
		},
		{
			name:     "correct permissions when user is creator of room",
			wantCode: http.StatusOK,
			wantBody: map[string]map[string]map[string]any{
				"permissions": {
					"room": {
						"canLock": true,
						"key":     id.String(),
					},
				},
			},
			request: "/api/estimation/room/1/bla/permissions",
		},
		{
			name:     "correct permissions when user is not creator of room",
			wantCode: http.StatusOK,
			wantBody: map[string]map[string]map[string]any{
				"permissions": {
					"room": {
						"canLock": false,
					},
				},
			},
			request: "/api/estimation/room/1/any-other/permissions",
		},
	}

	app := &application{
		rooms: map[internal.RoomId]*internal.Room{
			"1": {
				Id:            "1",
				NameOfCreator: "bla",
				Key:           id,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.request, nil)

			router.ServeHTTP(recorder, request)

			var got map[string]map[string]map[string]any
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotCode := recorder.Code

			assert.Equal(t, gotCode, test.wantCode)
			assert.DeepEqual(t, got, test.wantBody)
		})
	}
}
