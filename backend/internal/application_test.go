package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestApplication_handleRoundInRoomInProgress(t *testing.T) {
	tests := []struct {
		name        string
		expectation map[string]bool
		rooms       map[RoomId]*Room
		room        string
	}{
		{
			name: "not in progress when rooms are empty",
			expectation: map[string]bool{
				"inProgress": false,
				"isLocked":   false,
			},
			rooms: map[RoomId]*Room{},
			room:  "1",
		},
		{
			name: "in progress when room is set",
			expectation: map[string]bool{
				"inProgress": true,
				"isLocked":   false,
			},
			rooms: map[RoomId]*Room{
				"1": {
					inProgress: true,
				},
			},
			room: "1",
		},
	}

	for _, test := range tests {
		app := &Application{
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
			if gotContentType != "application/json" {
				t.Errorf("expected content type application/json, got %v", gotContentType)
			}

			if !reflect.DeepEqual(test.expectation, got) {
				t.Errorf("expected %v, got %v", test.expectation, got)
			}
		})
	}
}

func TestApplication_handleUserInRoomExists(t *testing.T) {
	client := &Client{
		Name: "Test",
	}
	roomId := "1"
	tests := []struct {
		url               string
		name              string
		rooms             map[RoomId]*Room
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
			rooms: map[RoomId]*Room{
				RoomId(roomId): {
					clients: make(map[*Client]bool),
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
			rooms: map[RoomId]*Room{
				RoomId(roomId): {
					clients: map[*Client]bool{
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
			rooms: map[RoomId]*Room{
				RoomId(roomId): {
					clients: map[*Client]bool{
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
			rooms:             map[RoomId]*Room{},
			expectedErrorBody: nil,
			statusCode:        200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &Application{
				rooms: test.rooms,
			}
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.url, nil)

			router.ServeHTTP(recorder, request)

			var got map[string]interface{}
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")
			if gotContentType != "application/json" {
				t.Errorf("expected content type application/json, got %v", gotContentType)
			}

			if test.expectedErrorBody != nil && !reflect.DeepEqual(test.expectedErrorBody, got) {
				t.Errorf("expected error body %v, got %v", test.expectedErrorBody, got)
			}

			if test.expectation != nil && !reflect.DeepEqual(test.expectation, got) {
				t.Errorf("expected %v, got %v", test.expectation, got)
			}

			if recorder.Code != test.statusCode {
				t.Errorf("expected status code %d, got %d", test.statusCode, recorder.Code)
			}
		})
	}
}

func TestApplication_handleFetchUsers(t *testing.T) {
	room := newRoom("1", make(chan<- RoomId), "")
	dev := newClient("B", Developer, room, nil)
	otherDev := newClient("Another", Developer, room, nil)
	devWithEqualLetter := newClient("Also a dev", Developer, room, nil)
	productOwner := newClient("Another one", ProductOwner, room, nil)
	otherProductOwner := newClient("Also a po", ProductOwner, room, nil)

	tests := []struct {
		name        string
		roomId      string
		rooms       map[RoomId]*Room
		expectation map[string][]userDTO
	}{
		{
			name:   "some users in the same room",
			roomId: "1",
			rooms: map[RoomId]*Room{
				RoomId("1"): {
					id:         "1",
					inProgress: false,
					leave:      nil,
					join:       nil,
					clients: map[*Client]bool{
						dev:                true,
						otherDev:           true,
						devWithEqualLetter: true,
						productOwner:       true,
						otherProductOwner:  true,
					},
					broadcast: nil,
					destroy:   nil,
				},
			},
			expectation: map[string][]userDTO{
				"productOwnerList": {
					{
						"name": "Also a po",
						"role": ProductOwner,
					},
					{
						"name": "Another one",
						"role": ProductOwner,
					},
				},
				"developerList": {
					{
						"name":   "Also a dev",
						"isDone": false,
						"role":   Developer,
					},
					{
						"name":   "Another",
						"isDone": false,
						"role":   Developer,
					},
					{
						"name":   "B",
						"isDone": false,
						"role":   Developer,
					},
				},
			},
		},
		{
			name:   "no clients",
			roomId: "1",
			rooms:  make(map[RoomId]*Room),
			expectation: map[string][]userDTO{
				"productOwnerList": {},
				"developerList":    {},
			},
		},
		{
			name:   "one dev client",
			roomId: "1",
			rooms: map[RoomId]*Room{
				RoomId("1"): {
					clients: map[*Client]bool{
						dev: true,
					},
				},
			},
			expectation: map[string][]userDTO{
				"productOwnerList": {},
				"developerList": {
					{
						"name":   "B",
						"isDone": false,
						"role":   Developer,
					},
				},
			},
		},
		{
			name: "one po client",
			rooms: map[RoomId]*Room{
				RoomId("1"): {
					clients: map[*Client]bool{
						productOwner: true,
					},
				},
			},
			roomId: "1",
			expectation: map[string][]userDTO{
				"productOwnerList": {
					{
						"name": "Another one",
						"role": ProductOwner,
					},
				},
				"developerList": {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &Application{
				guessConfig: &GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: nil,
			}
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/estimation/room/%s/users", test.roomId), nil)

			router.ServeHTTP(recorder, request)

			var got map[string][]userDTO
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")
			if gotContentType != "application/json" {
				t.Errorf("expected content type application/json, got %v", gotContentType)
			}

			if !reflect.DeepEqual(test.expectation, got) {
				t.Errorf("expected %v, got %v", test.expectation, got)
			}
		})
	}
}

func TestApplication_handleWs(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		rooms          map[RoomId]*Room
		expectedError  map[string]string
		expectedRoomId string
		expectedRole   string
		expectedStatus int
	}{
		{
			name: "register as developer",
			url:  "/api/estimation/room/1/developer?name=Test",
			rooms: map[RoomId]*Room{
				RoomId("1"): {
					id:         "1",
					inProgress: false,
					leave:      nil,
					join:       make(chan *Client),
					clients:    nil,
					broadcast:  make(chan *message),
					destroy:    nil,
				},
			},
			expectedError:  nil,
			expectedStatus: -1,
			expectedRoomId: "1",
			expectedRole:   Developer,
		},
		{
			name: "register as product owner",
			url:  "/api/estimation/room/1/product-owner?name=Test",
			rooms: map[RoomId]*Room{
				RoomId("1"): {
					id:         "1",
					inProgress: false,
					leave:      nil,
					join:       make(chan *Client),
					clients:    nil,
					broadcast:  make(chan *message),
					destroy:    nil,
				},
			},
			expectedError:  nil,
			expectedStatus: -1,
			expectedRoomId: "1",
			expectedRole:   ProductOwner,
		},
		{
			name:  "not registering due to missing name",
			url:   "/api/estimation/room/1/product-owner?name=",
			rooms: make(map[RoomId]*Room),
			expectedError: map[string]string{
				"message": "name is missing in query",
			},
			expectedStatus: 400,
			expectedRoomId: "1",
			expectedRole:   ProductOwner,
		},
		{
			name:  "not registering due to name too long",
			url:   "/api/estimation/room/1/product-owner?name=mynameiswaytoolongitshouldnotbecreated",
			rooms: make(map[RoomId]*Room),
			expectedError: map[string]string{
				"message": "name and room must be smaller or equal to 15",
			},
			expectedStatus: 400,
			expectedRoomId: "1",
			expectedRole:   ProductOwner,
		},
		{
			name:  "not registering due to roomId too long",
			url:   "/api/estimation/room/whateverthatroomisitiswaytoolong/product-owner?name=nameok",
			rooms: make(map[RoomId]*Room),
			expectedError: map[string]string{
				"message": "name and room must be smaller or equal to 15",
			},
			expectedStatus: 400,
			expectedRoomId: "1",
			expectedRole:   ProductOwner,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedMsg := newJoin()
			app := &Application{
				guessConfig: &GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: make(chan RoomId),
			}
			router := app.Routes()

			server := httptest.NewServer(router)
			defer server.Close()

			url := "ws" + strings.TrimPrefix(server.URL, "http") + test.url
			_, response, _ := websocket.Dial(context.Background(), url, nil)

			if test.expectedError != nil {
				var got map[string]string
				json.NewDecoder(response.Body).Decode(&got)
				if !reflect.DeepEqual(test.expectedError, got) {
					t.Errorf("expected response error %v, got %v", test.expectedError, got)
				}
				return
			}

			registeredClient := <-app.rooms[RoomId(test.expectedRoomId)].join
			broadcastedMsg := <-app.rooms[RoomId(test.expectedRoomId)].broadcast

			if registeredClient.Role != test.expectedRole {
				t.Errorf("expected to register client with role %s, got %v", test.expectedRole, registeredClient.Role)
			}

			if !reflect.DeepEqual(expectedMsg, broadcastedMsg) {
				t.Errorf("expected msg %v to broadcast, got %v", expectedMsg, broadcastedMsg)
			}
		})
	}
}

func TestApplication_handleWs_CreatingNewRoom(t *testing.T) {
	app := NewApplication(&GuessConfig{}, nil)
	roomId := "Test"
	expectedRoom := newRoom(RoomId(roomId), app.destroyRoom, "")
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
	got := app.rooms[RoomId(roomId)]
	app.roomMu.Unlock()

	if expectedRoom.id != got.id {
		t.Errorf("want room with id %v, got %v", expectedRoom.id, got.id)
	}
}

func TestApplication_handleFetchActiveRooms(t *testing.T) {
	tests := []struct {
		name  string
		rooms map[RoomId]*Room
		want  []string
	}{
		{
			name: "with multiple active rooms",
			rooms: map[RoomId]*Room{
				RoomId("Blub"): {
					id: "Blub",
				},
				RoomId("Test"): {
					id: "Test",
				},
			},
			want: []string{"Blub", "Test"},
		},
		{
			name:  "no rooms",
			rooms: make(map[RoomId]*Room),
			want:  []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := &Application{
				guessConfig: &GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: nil,
			}
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/estimation/room/rooms", nil)

			router.ServeHTTP(recorder, request)

			var got []string
			json.Unmarshal(recorder.Body.Bytes(), &got)

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func TestApplication_handlePossibleGuesses(t *testing.T) {
	tests := []struct {
		name   string
		config *GuessConfig
		want   []map[string]interface{}
	}{
		{
			name: "multiple guesses",
			config: &GuessConfig{
				Guesses: []guessConfigEntry{
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
			config: &GuessConfig{
				Guesses: []guessConfigEntry{
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
			config: &GuessConfig{
				Guesses: make([]guessConfigEntry, 0),
			},
			want: make([]map[string]interface{}, 0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := NewApplication(test.config, nil)
			router := app.Routes()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/estimation/possible-guesses", nil)

			router.ServeHTTP(recorder, request)

			var got []map[string]interface{}
			json.Unmarshal(recorder.Body.Bytes(), &got)

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func TestApplication_ListenForRoomDestroy(t *testing.T) {
	destroyChannel := make(chan RoomId)
	roomToDestroy := RoomId("Test")
	app := &Application{
		guessConfig: &GuessConfig{},
		rooms: map[RoomId]*Room{
			roomToDestroy: {
				id:         roomToDestroy,
				inProgress: false,
				leave:      nil,
				join:       nil,
				clients:    nil,
				broadcast:  nil,
				destroy:    nil,
			},
		},
		destroyRoom: destroyChannel,
	}
	go app.ListenForRoomDestroy()

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
	app := &Application{
		guessConfig: &GuessConfig{},
		rooms: map[RoomId]*Room{
			"1": {
				id:             "1",
				inProgress:     false,
				leave:          nil,
				join:           nil,
				clients:        nil,
				broadcast:      nil,
				destroy:        nil,
				hashedPassword: hashedPassword,
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

			if gotCode != test.wantCode {
				t.Errorf("want %v, got %v", test.wantCode, gotCode)
			}

			if !reflect.DeepEqual(test.wantBody, got) {
				t.Errorf("want %v, got %v", test.wantBody, got)
			}
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

	app := &Application{
		guessConfig: &GuessConfig{},
		rooms: map[RoomId]*Room{
			"1": {
				id:            "1",
				nameOfCreator: "bla",
				key:           id,
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

			if gotCode != test.wantCode {
				t.Errorf("want %v, got %v", test.wantCode, gotCode)
			}

			if !reflect.DeepEqual(test.wantBody, got) {
				t.Errorf("want %#v, got %#v", test.wantBody, got)
			}
		})
	}
}
