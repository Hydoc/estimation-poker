package main

import (
	"context"
	"net/http"
	"sort"

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
	room := internal.NewRoom(roomId, app.destroyRoom, input.Creator, app.logger, app.guessConfig)
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

	room, ok := app.rooms[roomId]

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

	actualRoom, ok := app.rooms[roomId]
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

	actualRoom, ok := app.rooms[roomId]
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
