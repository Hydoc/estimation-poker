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
	"testing"
)

func TestApplication_handleRoundInRoomInProgress(t *testing.T) {
	testSuites := []struct {
		name        string
		expectation map[string]bool
		rooms       map[string]bool
		room        string
	}{
		{
			name: "not in progress when rooms are empty",
			expectation: map[string]bool{
				"inProgress": false,
			},
			rooms: map[string]bool{},
			room:  "1",
		},
		{
			name: "in progress when room is set",
			expectation: map[string]bool{
				"inProgress": true,
			},
			rooms: map[string]bool{
				"1": true,
			},
			room: "1",
		},
	}

	for _, suite := range testSuites {
		hub := &Hub{
			roomBroadcast: make(chan roomBroadcastMessage),
			register:      make(chan *Client),
			unregister:    make(chan *Client),
			clients:       make(map[*Client]bool),
			rooms:         suite.rooms,
		}

		t.Run(suite.name, func(t *testing.T) {
			upgrdr := &websocket.Upgrader{}
			app := NewApplication(mux.NewRouter(), upgrdr, hub)
			router := app.ConfigureRouting()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/estimation/room/%s/state", suite.room), nil)

			router.ServeHTTP(recorder, request)

			var got map[string]bool
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")
			if gotContentType != "application/json" {
				t.Errorf("expected content type application/json, got %v", gotContentType)
			}

			if !reflect.DeepEqual(suite.expectation, got) {
				t.Errorf("expected %v, got %v", suite.expectation, got)
			}
		})
	}
}

func TestApplication_handleUserInRoomExists(t *testing.T) {
	client := &Client{
		RoomId: "1",
		Name:   "Test",
	}
	testSuites := []struct {
		url               string
		name              string
		expectation       map[string]interface{}
		clients           map[*Client]bool
		expectedErrorBody map[string]interface{}
		statusCode        int
	}{
		{
			name: "not find client when clients are empty",
			url:  "/api/estimation/room/1/users/exists?name=Bla",
			expectation: map[string]interface{}{
				"exists": false,
			},
			clients:    make(map[*Client]bool),
			statusCode: 200,
		},
		{
			name: "find client when client exists",
			url:  fmt.Sprintf("/api/estimation/room/1/users/exists?name=%s", client.Name),
			expectation: map[string]interface{}{
				"exists": true,
			},
			clients: map[*Client]bool{
				client: true,
			},
			statusCode: 409,
		},
		{
			name:        "error when trying to access without name",
			url:         "/api/estimation/room/1/users/exists?name=",
			expectation: nil,
			clients: map[*Client]bool{
				client: true,
			},
			expectedErrorBody: map[string]interface{}{
				"message": "name is missing in query",
			},
			statusCode: 400,
		},
	}

	for _, suite := range testSuites {
		hub := &Hub{
			roomBroadcast: make(chan roomBroadcastMessage),
			register:      make(chan *Client),
			unregister:    make(chan *Client),
			clients:       suite.clients,
			rooms:         make(map[string]bool),
		}

		t.Run(suite.name, func(t *testing.T) {
			upgrdr := &websocket.Upgrader{}
			app := NewApplication(mux.NewRouter(), upgrdr, hub)
			router := app.ConfigureRouting()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, suite.url, nil)

			router.ServeHTTP(recorder, request)

			var got map[string]interface{}
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")
			if gotContentType != "application/json" {
				t.Errorf("expected content type application/json, got %v", gotContentType)
			}

			if suite.expectedErrorBody != nil && !reflect.DeepEqual(suite.expectedErrorBody, got) {
				t.Errorf("expected error body %v, got %v", suite.expectedErrorBody, got)
			}

			if suite.expectation != nil && !reflect.DeepEqual(suite.expectation, got) {
				t.Errorf("expected %v, got %v", suite.expectation, got)
			}

			if recorder.Code != suite.statusCode {
				t.Errorf("expected status code %d, got %d", suite.statusCode, recorder.Code)
			}
		})
	}
}

func TestApplication_handleFetchUsers(t *testing.T) {
	dev := newDeveloper("1", "B", nil, nil)
	otherDev := newDeveloper("1", "Another", nil, nil)
	devWithEqualLetter := newDeveloper("1", "Also a dev", nil, nil)
	productOwner := newProductOwner("1", "Another one", nil, nil)
	otherProductOwner := newProductOwner("1", "Also a po", nil, nil)
	productOwnerInDifferentRoom := &Client{Name: "Different Room", Role: ProductOwner, Guess: 0, RoomId: "different Room"}

	testSuites := []struct {
		name        string
		clients     map[*Client]bool
		roomId      string
		expectation map[string][]userDTO
	}{
		{
			name:   "some users in the same room",
			roomId: "1",
			clients: map[*Client]bool{
				dev:                         true,
				otherDev:                    true,
				devWithEqualLetter:          true,
				productOwner:                true,
				otherProductOwner:           true,
				productOwnerInDifferentRoom: true,
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
			name:    "no clients",
			roomId:  "1",
			clients: map[*Client]bool{},
			expectation: map[string][]userDTO{
				"productOwnerList": {},
				"developerList":    {},
			},
		},
		{
			name:   "one dev client",
			roomId: "1",
			clients: map[*Client]bool{
				dev: true,
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
			name:   "one po client",
			roomId: "1",
			clients: map[*Client]bool{
				productOwner: true,
			},
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

	for _, suite := range testSuites {
		t.Run(suite.name, func(t *testing.T) {
			hub := &Hub{
				roomBroadcast: make(chan roomBroadcastMessage),
				register:      make(chan *Client),
				unregister:    make(chan *Client),
				clients:       suite.clients,
				rooms:         make(map[string]bool),
			}
			upgrdr := &websocket.Upgrader{}
			app := NewApplication(mux.NewRouter(), upgrdr, hub)
			router := app.ConfigureRouting()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/estimation/room/%s/users", suite.roomId), nil)

			router.ServeHTTP(recorder, request)

			var got map[string][]userDTO
			json.Unmarshal(recorder.Body.Bytes(), &got)

			gotContentType := recorder.Header().Get("Content-Type")
			if gotContentType != "application/json" {
				t.Errorf("expected content type application/json, got %v", gotContentType)
			}

			if !reflect.DeepEqual(suite.expectation, got) {
				t.Errorf("expected %v, got %v", suite.expectation, got)
			}
		})
	}
}

func TestApplication_handleWs(t *testing.T) {
	testSuites := []struct {
		name           string
		url            string
		expectedError  map[string]string
		expectedRoomId string
		expectedRole   string
		expectedStatus int
	}{
		{
			name:           "register as developer",
			url:            "/api/estimation/room/1/developer?name=Test",
			expectedError:  nil,
			expectedStatus: -1,
			expectedRoomId: "1",
			expectedRole:   Developer,
		},
		{
			name:           "register as product owner",
			url:            "/api/estimation/room/1/product-owner?name=Test",
			expectedError:  nil,
			expectedStatus: -1,
			expectedRoomId: "1",
			expectedRole:   ProductOwner,
		},
		{
			name: "not registering due to missing name",
			url:  "/api/estimation/room/1/product-owner?name=",
			expectedError: map[string]string{
				"message": "name is missing in query",
			},
			expectedStatus: 400,
			expectedRoomId: "1",
			expectedRole:   ProductOwner,
		},
	}

	for _, suite := range testSuites {
		t.Run(suite.name, func(t *testing.T) {
			expectedMsg := newJoin()
			hub := &Hub{
				roomBroadcast: make(chan roomBroadcastMessage),
				register:      make(chan *Client),
				unregister:    make(chan *Client),
				clients:       make(map[*Client]bool),
				rooms:         make(map[string]bool),
			}
			upgrdr := &websocket.Upgrader{}
			app := NewApplication(mux.NewRouter(), upgrdr, hub)
			router := app.ConfigureRouting()

			server := httptest.NewServer(router)
			defer server.Close()

			url := "ws" + strings.TrimPrefix(server.URL, "http") + suite.url
			_, response, _ := websocket.DefaultDialer.Dial(url, nil)

			if suite.expectedError != nil {
				var got map[string]string
				json.NewDecoder(response.Body).Decode(&got)
				if !reflect.DeepEqual(suite.expectedError, got) {
					t.Errorf("expected response error %v, got %v", suite.expectedError, got)
				}
				return
			}

			registeredClient := <-hub.register
			broadcastedMsg := <-hub.roomBroadcast

			if registeredClient.Role != suite.expectedRole {
				t.Errorf("expected to register client with role %s, got %v", suite.expectedRole, registeredClient.Role)
			}

			if !reflect.DeepEqual(expectedMsg, broadcastedMsg.message) {
				t.Errorf("expected msg %v to broadcast, got %v", expectedMsg, broadcastedMsg.message)
			}

			if suite.expectedRoomId != broadcastedMsg.RoomId {
				t.Errorf("expected roomId %v, got %v", suite.expectedRoomId, broadcastedMsg.RoomId)
			}
		})
	}
}

func TestApplication_handleWs_UpgradingConnectionFailed(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	hub := &Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
	app := NewApplication(mux.NewRouter(), &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return false
	}}, hub)
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
		name    string
		clients map[*Client]bool
		want    []string
	}{
		{
			name: "with multiple active rooms",
			clients: map[*Client]bool{
				&Client{RoomId: "Blub"}: true,
				&Client{RoomId: "Blub"}: true,
				&Client{RoomId: "Test"}: true,
			},
			want: []string{"Blub", "Test"},
		},
		{
			name:    "no rooms",
			clients: map[*Client]bool{},
			want:    []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hub := &Hub{
				clients: test.clients,
			}

			app := NewApplication(mux.NewRouter(), &websocket.Upgrader{}, hub)
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
