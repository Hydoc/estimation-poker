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

	"github.com/Hydoc/estimation-poker/backend/internal"
)

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
	room := internal.NewRoom(internal.RoomId(roomId.String()), app.destroyRoom, input.Creator, app.logger, app.guessConfig)
	app.rooms[room.Id] = room
	go room.Run()

	err = app.writeJSON(writer, http.StatusCreated, envelope{"id": roomId.String()}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}

func (app *application) handleFetchRoomMetadata(writer http.ResponseWriter, request *http.Request) {
	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	room, ok := app.rooms[internal.RoomId(roomId.String())]

	if !ok {
		err = app.writeJSON(writer, http.StatusOK, envelope{"exists": false, "isLocked": false}, nil)
		if err != nil {
			app.serverErrorResponse(writer, request, err)
			return
		}
		return
	}

	err = app.writeJSON(writer, http.StatusOK, envelope{"exists": true, "isLocked": room.IsLocked()}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}
}

func (app *application) handleConnectionState(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	defer request.Body.Close()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err = app.readJSON(writer, request, &input)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
		return
	}

	actualRoom, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		app.notFoundResponse(writer, request)
		return
	}

	err = app.writeJSON(writer, http.StatusOK, actualRoom.ConnectionState(input.Username, input.Password), nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}

func (app *application) handleFetchRoomState(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	actualRoom, ok := app.rooms[internal.RoomId(roomId.String())]
	if !ok {
		app.notFoundResponse(writer, request)
		return
	}

	err = app.writeJSON(writer, http.StatusOK, actualRoom.State(), nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
}

func (app *application) handleFetchActiveRooms(writer http.ResponseWriter, request *http.Request) {
	//goland:noinspection GoPreferNilSlice
	overviewRooms := []internal.Overview{}
	for _, room := range app.rooms {
		if !room.IsLocked() {
			overviewRooms = append(overviewRooms, room.AsOverview())
		}
	}
	sort.Slice(overviewRooms, func(i, j int) bool {
		return overviewRooms[i].Created.Before(overviewRooms[j].Created)
	})
	err := app.writeJSON(writer, http.StatusOK, envelope{"rooms": overviewRooms}, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
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
		err = app.writeJSON(writer, http.StatusOK, []map[string]any{}, nil)
		if err != nil {
			app.serverErrorResponse(writer, request, err)
		}
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

	err = app.writeJSON(writer, http.StatusOK, out, nil)
	if err != nil {
		app.serverErrorResponse(writer, request, err)
	}
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

	go client.WebsocketReader()
	go client.WebsocketWriter()
	clientRoom.Join(client)
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
