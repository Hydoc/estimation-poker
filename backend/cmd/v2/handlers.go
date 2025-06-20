package main

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/Hydoc/guess-dev/backend/internal/game"
)

func (app *application) healthcheckHandler(writer http.ResponseWriter, _ *http.Request) {
	app.writeJSON(writer, http.StatusOK, envelope{"status": "ok"}, nil)
}

func (app *application) handleWS(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		app.writeJSON(writer, http.StatusBadRequest, envelope{"error": err.Error()}, nil)
		return
	}

	id, err := app.readUUIDParam(request)
	if err != nil {
		app.writeJSON(writer, http.StatusBadRequest, envelope{"error": err.Error()}, nil)
		return
	}

	room := app.rooms.Find(id)
	if room == nil {
		app.writeJSON(writer, http.StatusNotFound, envelope{"error": "room not found"}, nil)
		return
	}

	conn, err := websocket.Accept(writer, request, nil)
	if err != nil {
		app.writeJSON(writer, http.StatusBadRequest, envelope{"error": err.Error()}, nil)
		return
	}
	defer conn.CloseNow()
	app.writeJSON(writer, http.StatusOK, envelope{"room": room.Id}, nil)
}

func (app *application) roomHandler(writer http.ResponseWriter, _ *http.Request) {
	id := uuid.New()
	app.rooms.Save(game.NewRoom(id))
	app.writeJSON(writer, http.StatusOK, envelope{"id": id.String()}, nil)
}
