package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
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
			},
			rooms: map[RoomId]*Room{},
			room:  "1",
		},
		{
			name: "in progress when room is set",
			expectation: map[string]bool{
				"inProgress": true,
			},
			rooms: map[RoomId]*Room{
				"1": {
					InProgress: true,
				},
			},
			room: "1",
		},
	}

	for _, test := range tests {
		app := &Application{
			rooms:  test.rooms,
			router: mux.NewRouter(),
		}

		t.Run(test.name, func(t *testing.T) {
			router := app.ConfigureRouting()
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
				router: mux.NewRouter(),
				rooms:  test.rooms,
			}
			router := app.ConfigureRouting()
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
	room := newRoom("1", make(chan<- RoomId))
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
					InProgress: false,
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
						"name":  "Also a dev",
						"guess": float64(0),
						"role":  Developer,
					},
					{
						"name":  "Another",
						"guess": float64(0),
						"role":  Developer,
					},
					{
						"name":  "B",
						"guess": float64(0),
						"role":  Developer,
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
						"name":  "B",
						"guess": float64(0),
						"role":  Developer,
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
				router:      mux.NewRouter(),
				upgrader:    &websocket.Upgrader{},
				guessConfig: &GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: nil,
			}
			router := app.ConfigureRouting()
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
					InProgress: false,
					leave:      nil,
					join:       make(chan *Client),
					clients:    nil,
					broadcast:  make(chan message),
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
					InProgress: false,
					leave:      nil,
					join:       make(chan *Client),
					clients:    nil,
					broadcast:  make(chan message),
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedMsg := newJoin()
			app := &Application{
				router:      mux.NewRouter(),
				upgrader:    &websocket.Upgrader{},
				guessConfig: &GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: make(chan RoomId),
			}
			router := app.ConfigureRouting()

			server := httptest.NewServer(router)
			defer server.Close()

			url := "ws" + strings.TrimPrefix(server.URL, "http") + test.url
			_, response, _ := websocket.DefaultDialer.Dial(url, nil)

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
	app := NewApplication(mux.NewRouter(), &websocket.Upgrader{}, &GuessConfig{})
	roomId := "Test"
	expectedRoom := newRoom(RoomId(roomId), app.destroyRoom)
	router := app.ConfigureRouting()

	server := httptest.NewServer(router)
	defer server.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		url := "ws" + strings.TrimPrefix(server.URL, "http") + fmt.Sprintf("/api/estimation/room/%s/developer?name=Test", roomId)
		websocket.DefaultDialer.Dial(url, nil)
	}()

	wg.Wait()

	got := app.rooms[RoomId(roomId)]

	if expectedRoom.id != got.id {
		t.Errorf("want room with id %v, got %v", expectedRoom.id, got.id)
	}
}

func TestApplication_handleWs_UpgradingConnectionFailed(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	app := NewApplication(mux.NewRouter(), &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return false
	}}, &GuessConfig{})
	router := app.ConfigureRouting()

	server := httptest.NewServer(router)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/estimation/room/1/product-owner?name=Test"
	websocket.DefaultDialer.Dial(url, nil)

	wantLog := "upgrade: websocket: request origin not allowed"
	if !strings.Contains(logBuffer.String(), wantLog) {
		t.Errorf("expected to log %v, got %v", wantLog, logBuffer.String())
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
				router:      mux.NewRouter(),
				upgrader:    &websocket.Upgrader{},
				guessConfig: &GuessConfig{},
				rooms:       test.rooms,
				destroyRoom: nil,
			}
			router := app.ConfigureRouting()
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
			app := NewApplication(mux.NewRouter(), &websocket.Upgrader{}, test.config)
			router := app.ConfigureRouting()
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
		router:      mux.NewRouter(),
		upgrader:    &websocket.Upgrader{},
		guessConfig: &GuessConfig{},
		rooms: map[RoomId]*Room{
			roomToDestroy: {
				id:         roomToDestroy,
				InProgress: false,
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

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.destroyRoom <- roomToDestroy
	}()

	wg.Wait()

	if _, ok := app.rooms[roomToDestroy]; ok {
		t.Error("expected app to not have room")
	}
}
