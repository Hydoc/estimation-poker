package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/coder/websocket"
)

type Application struct {
	roomMu sync.Mutex

	logger      *slog.Logger
	guessConfig *GuessConfig
	rooms       map[RoomId]*Room
	destroyRoom chan RoomId
}

type envelope map[string]any

func (app *Application) withRequiredQueryParam(param string, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		queryParam := request.URL.Query().Get(param)

		if len(queryParam) == 0 || !request.URL.Query().Has(param) {
			app.writeJson(writer, http.StatusBadRequest, envelope{"message": fmt.Sprintf("%s is missing in query", param)}, nil)
			return
		}

		next.ServeHTTP(writer, request)
	}
}

func (app *Application) handleRoomAuthenticate(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	defer request.Body.Close()
	roomId := request.PathValue("id")
	actualRoom, ok := app.rooms[RoomId(roomId)]
	if !ok {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	var input struct {
		Password string `json:"password"`
	}

	err := json.NewDecoder(request.Body).Decode(&input)
	if err != nil {
		app.writeJson(writer, http.StatusOK, envelope{"ok": false}, nil)
		return
	}

	if actualRoom.verify(input.Password) {
		app.writeJson(writer, http.StatusOK, envelope{"ok": true}, nil)
		return
	}

	app.writeJson(writer, http.StatusOK, envelope{"ok": false}, nil)
}

func (app *Application) handleFetchPermissions(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	roomId := request.PathValue("id")
	username := request.PathValue("username")
	actualRoom, ok := app.rooms[RoomId(roomId)]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if actualRoom.nameOfCreator == username {
		app.writeJson(writer, http.StatusOK, map[string]map[string]map[string]any{
			"permissions": {
				"room": {
					"canLock": true,
					"key":     actualRoom.key.String(),
				},
			},
		}, nil)
		return
	}

	app.writeJson(writer, http.StatusOK, map[string]map[string]map[string]any{
		"permissions": {
			"room": {
				"canLock": false,
			},
		},
	}, nil)
}

func (app *Application) handlePossibleGuesses(writer http.ResponseWriter, _ *http.Request) {
	app.writeJson(writer, http.StatusOK, app.guessConfig.Guesses, nil)
}

func (app *Application) handleFetchRoomState(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	roomId := request.PathValue("id")
	actualRoom, ok := app.rooms[RoomId(roomId)]
	if !ok {
		app.writeJson(writer, http.StatusOK, envelope{
			"inProgress": false,
			"isLocked":   false,
		}, nil)
		return
	}
	app.writeJson(writer, http.StatusOK, envelope{
		"inProgress": actualRoom.inProgress,
		"isLocked":   actualRoom.isLocked,
	}, nil)
}

func (app *Application) handleUserInRoomExists(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()
	roomId := request.PathValue("id")

	name := request.URL.Query().Get("name")

	if _, ok := app.rooms[RoomId(roomId)]; !ok {
		app.writeJson(writer, http.StatusOK, envelope{"exists": false}, nil)
		return
	}

	for client := range app.rooms[RoomId(roomId)].clients {
		if client.Name == name {
			app.writeJson(writer, http.StatusConflict, envelope{"exists": true}, nil)
			return
		}
	}
	app.writeJson(writer, http.StatusOK, envelope{"exists": false}, nil)
}

func (app *Application) handleFetchActiveRooms(writer http.ResponseWriter, _ *http.Request) {
	activeRooms := []string{}
	for _, room := range app.rooms {
		if !room.isLocked {
			activeRooms = append(activeRooms, string(room.id))
		}
	}
	slices.Sort(activeRooms)
	app.writeJson(writer, http.StatusOK, activeRooms, nil)
}

func (app *Application) handleFetchUsers(writer http.ResponseWriter, request *http.Request) {
	roomId := request.PathValue("id")

	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	var usersInRoom = map[string][]userDTO{
		"productOwnerList": {},
		"developerList":    {},
	}
	var clients []*Client

	if _, ok := app.rooms[RoomId(roomId)]; !ok {
		app.writeJson(writer, http.StatusOK, usersInRoom, nil)
		return
	}

	for client := range app.rooms[RoomId(roomId)].clients {
		clients = append(clients, client)
	}
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].Name < clients[j].Name
	})

	for _, c := range clients {
		switch c.Role {
		case Developer:
			usersInRoom["developerList"] = append(usersInRoom["developerList"], c.toJson())
		case ProductOwner:
			usersInRoom["productOwnerList"] = append(usersInRoom["productOwnerList"], c.toJson())
		}
	}
	app.writeJson(writer, http.StatusOK, usersInRoom, nil)
}

func (app *Application) handleWs(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	roomId := request.PathValue("id")

	name := request.URL.Query().Get("name")

	if utf8.RuneCountInString(name) > 15 || utf8.RuneCountInString(roomId) > 15 {
		app.writeJson(writer, http.StatusBadRequest, envelope{"message": "name and room must be smaller or equal to 15"}, nil)
		return
	}

	connection, err := websocket.Accept(writer, request, nil)
	if err != nil {
		app.logger.Info(fmt.Sprintf("upgrade: %s", err))
		return
	}

	var clientRoom *Room
	if room, ok := app.rooms[RoomId(roomId)]; ok {
		// when there is a room, it's already running in a goroutine
		clientRoom = room
	} else {
		clientRoom = newRoom(RoomId(roomId), app.destroyRoom, name)
		app.rooms[clientRoom.id] = clientRoom
		go clientRoom.Run()
	}

	clientRole := Developer
	if strings.Contains(request.URL.Path, "product-owner") {
		clientRole = ProductOwner
	}
	client := newClient(name, clientRole, clientRoom, connection)
	clientRoom.join <- client
	clientRoom.broadcast <- newJoin()

	go client.websocketReader()
	go client.websocketWriter()
}

func (app *Application) ListenForRoomDestroy() {
	for {
		select {
		case roomId := <-app.destroyRoom:
			app.roomMu.Lock()
			if _, ok := app.rooms[roomId]; ok {
				delete(app.rooms, roomId)
			}
			app.roomMu.Unlock()
		}
	}
}

func NewApplication(config *GuessConfig, logger *slog.Logger) *Application {
	return &Application{
		logger:      logger,
		guessConfig: config,
		rooms:       make(map[RoomId]*Room),
		destroyRoom: make(chan RoomId),
	}
}
