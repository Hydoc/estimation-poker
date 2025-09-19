package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/Hydoc/guess-dev/backend/internal"
)

type application struct {
	roomMu sync.Mutex

	logger      *slog.Logger
	guessConfig *internal.GuessConfig
	rooms       map[internal.RoomId]*internal.Room
	destroyRoom chan internal.RoomId
}

type envelope map[string]any

func (app *application) withRequiredQueryParam(param string, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		queryParam := request.URL.Query().Get(param)

		if len(queryParam) == 0 || !request.URL.Query().Has(param) {
			app.writeJson(writer, http.StatusBadRequest, envelope{"message": fmt.Sprintf("%s is missing in query", param)}, nil)
			return
		}

		next.ServeHTTP(writer, request)
	}
}

func (app *application) handleRoomAuthenticate(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	defer request.Body.Close()
	roomId := request.PathValue("id")
	actualRoom, ok := app.rooms[internal.RoomId(roomId)]
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

	if actualRoom.Verify(input.Password) {
		app.writeJson(writer, http.StatusOK, envelope{"ok": true}, nil)
		return
	}

	app.writeJson(writer, http.StatusOK, envelope{"ok": false}, nil)
}

func (app *application) handleFetchPermissions(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	roomId := request.PathValue("id")
	username := request.PathValue("username")
	actualRoom, ok := app.rooms[internal.RoomId(roomId)]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if actualRoom.NameOfCreator == username {
		app.writeJson(writer, http.StatusOK, map[string]map[string]map[string]any{
			"permissions": {
				"room": {
					"canLock": true,
					"key":     actualRoom.Key.String(),
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

func (app *application) createNewRoom(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	name := request.URL.Query().Get("name")

	roomId := uuid.New()
	room := internal.NewRoom(internal.RoomId(roomId.String()), app.destroyRoom, name)
	app.rooms[room.Id] = room
	go room.Run()

	app.writeJson(writer, http.StatusCreated, envelope{"id": roomId.String()}, nil)
}

func (app *application) handlePossibleGuesses(writer http.ResponseWriter, _ *http.Request) {
	app.writeJson(writer, http.StatusOK, app.guessConfig.Guesses, nil)
}

func (app *application) handleFetchRoomState(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	roomId := request.PathValue("id")
	actualRoom, ok := app.rooms[internal.RoomId(roomId)]
	if !ok {
		app.writeJson(writer, http.StatusOK, envelope{
			"inProgress": false,
			"isLocked":   false,
		}, nil)
		return
	}
	app.writeJson(writer, http.StatusOK, envelope{
		"inProgress": actualRoom.InProgress,
		"isLocked":   actualRoom.IsLocked,
	}, nil)
}

func (app *application) handleUserInRoomExists(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()
	roomId := request.PathValue("id")

	name := request.URL.Query().Get("name")

	if _, ok := app.rooms[internal.RoomId(roomId)]; !ok {
		app.writeJson(writer, http.StatusOK, envelope{"exists": false}, nil)
		return
	}

	for client := range app.rooms[internal.RoomId(roomId)].Clients {
		if client.Name == name {
			app.writeJson(writer, http.StatusConflict, envelope{"exists": true}, nil)
			return
		}
	}
	app.writeJson(writer, http.StatusOK, envelope{"exists": false}, nil)
}

func (app *application) handleFetchActiveRooms(writer http.ResponseWriter, _ *http.Request) {
	var activeRooms []*internal.Room
	for _, room := range app.rooms {
		if !room.IsLocked {
			activeRooms = append(activeRooms, room)
		}
	}
	sort.Slice(activeRooms, func(i, j int) bool {
		return activeRooms[i].Created.Before(activeRooms[j].Created)
	})
	app.writeJson(writer, http.StatusOK, envelope{"rooms": activeRooms}, nil)
}

func (app *application) handleFetchUsers(writer http.ResponseWriter, request *http.Request) {
	roomId := request.PathValue("id")

	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	room, ok := app.rooms[internal.RoomId(roomId)]
	if !ok {
		app.writeJson(writer, http.StatusOK, []map[string]any{}, nil)
		return
	}

	var clients []*internal.Client
	for client := range room.Clients {
		clients = append(clients, client)
	}

	sort.Slice(clients, func(i, j int) bool {
		return clients[i].Name < clients[j].Name
	})

	var out []map[string]any
	for _, client := range clients {
		out = append(out, client.ToJson())
	}

	app.writeJson(writer, http.StatusOK, out, nil)
}

func (app *application) handleWs(writer http.ResponseWriter, request *http.Request) {
	app.roomMu.Lock()
	defer app.roomMu.Unlock()

	roomId, err := uuid.Parse(request.PathValue("id"))
	if err != nil {
		app.writeJson(writer, http.StatusBadRequest, envelope{"message": "roomId is invalid"}, nil)
		return
	}

	name := request.URL.Query().Get("name")

	if utf8.RuneCountInString(name) > 15 {
		app.writeJson(writer, http.StatusBadRequest, envelope{"message": "name must be smaller or equal to 15"}, nil)
		return
	}

	clientRoom, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		app.writeJson(writer, http.StatusNotFound, envelope{"message": "room not found"}, nil)
		return
	}

	connection, err := websocket.Accept(writer, request, nil)
	if err != nil {
		app.logger.Info(fmt.Sprintf("upgrade: %s", err))
		return
	}

	clientRole := internal.Developer
	if strings.Contains(request.URL.Path, "product-owner") {
		clientRole = internal.ProductOwner
	}
	client := internal.NewClient(name, clientRole, clientRoom, connection)
	clientRoom.Join <- client
	clientRoom.Broadcast <- internal.NewJoin()

	go client.WebsocketReader()
	go client.WebsocketWriter()
}

func (app *application) listenForRoomDestroy(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case roomId := <-app.destroyRoom:
			app.roomMu.Lock()
			delete(app.rooms, roomId)
			app.roomMu.Unlock()
		}
	}
}
