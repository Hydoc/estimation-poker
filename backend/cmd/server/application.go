package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/Hydoc/guess-dev/backend/internal"
)

func (app *application) handleRoomAuthenticate(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	defer request.Body.Close()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	actualRoom, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	var input struct {
		Password string `json:"password"`
	}

	err = app.readJSON(writer, request, &input)
	if err != nil {
		app.writeJSON(writer, http.StatusOK, envelope{"ok": false}, nil)
		return
	}

	if actualRoom.Verify(input.Password) {
		app.writeJSON(writer, http.StatusOK, envelope{"ok": true}, nil)
		return
	}

	app.writeJSON(writer, http.StatusOK, envelope{"ok": false}, nil)
}

func (app *application) handleFetchPermissions(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	username := request.URL.Query().Get("name")

	actualRoom, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		app.notFoundResponse(writer, request)
		return
	}

	if actualRoom.NameOfCreator == username {
		app.writeJSON(writer, http.StatusOK, map[string]map[string]map[string]any{
			"permissions": {
				"room": {
					"canLock": true,
					"key":     actualRoom.Key.String(),
				},
			},
		}, nil)
		return
	}

	app.writeJSON(writer, http.StatusOK, map[string]map[string]map[string]any{
		"permissions": {
			"room": {
				"canLock": false,
			},
		},
	}, nil)
}

func (app *application) createNewRoom(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	var input struct {
		Creator string         `json:"creator"`
		Guesses map[int]string `json:"guesses"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	roomId := uuid.New()
	room := internal.NewRoom(internal.RoomId(roomId.String()), app.destroyRoom, input.Creator, app.logger)
	app.rooms[room.Id] = room
	go room.Run()

	app.writeJSON(writer, http.StatusCreated, envelope{"id": roomId.String()}, nil)
}

func (app *application) handlePossibleGuesses(writer http.ResponseWriter, _ *http.Request) {
	app.writeJSON(writer, http.StatusOK, app.guessConfig.Guesses, nil)
}

func (app *application) handleRoomExists(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()
	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}
	_, ok := app.rooms[internal.RoomId(roomId.String())]
	app.writeJSON(writer, http.StatusOK, envelope{"exists": ok}, nil)
}

func (app *application) handleFetchRoomState(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.writeJSON(writer, http.StatusOK, envelope{
			"inProgress": false,
			"isLocked":   false,
		}, nil)
		return
	}

	actualRoom, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		app.writeJSON(writer, http.StatusOK, envelope{
			"inProgress": false,
			"isLocked":   false,
		}, nil)
		return
	}
	app.writeJSON(writer, http.StatusOK, envelope{
		"inProgress": actualRoom.InProgress,
		"isLocked":   actualRoom.IsLocked,
	}, nil)
}

func (app *application) handleUserInRoomExists(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	name := request.URL.Query().Get("name")

	if _, ok := app.rooms[internal.RoomId(roomId.String())]; !ok {
		app.writeJSON(writer, http.StatusOK, envelope{"exists": false}, nil)
		return
	}

	for client := range app.rooms[internal.RoomId(roomId.String())].Clients {
		if client.Name == name {
			app.writeJSON(writer, http.StatusConflict, envelope{"exists": true}, nil)
			return
		}
	}
	app.writeJSON(writer, http.StatusOK, envelope{"exists": false}, nil)
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
	app.writeJSON(writer, http.StatusOK, envelope{"rooms": activeRooms}, nil)
}

func (app *application) handleFetchUsers(writer http.ResponseWriter, request *http.Request) {
	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	room, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		app.writeJSON(writer, http.StatusOK, []map[string]any{}, nil)
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

	app.writeJSON(writer, http.StatusOK, out, nil)
}

func (app *application) handleWs(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	name := request.URL.Query().Get("name")

	if utf8.RuneCountInString(name) > 15 {
		app.badRequestResponse(writer, request, errors.New("name must be smaller or equal to 15"))
		return
	}

	clientRoom, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		app.notFoundResponse(writer, request)
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
	client := internal.NewClient(name, clientRole, clientRoom, connection, app.bus, app.logger)
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
			app.mu.Lock()
			delete(app.rooms, roomId)
			app.mu.Unlock()
		}
	}
}
