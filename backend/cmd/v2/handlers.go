package main

import "net/http"

func (app *application) healthcheckHandler(writer http.ResponseWriter, _ *http.Request) {
	app.writeJSON(writer, http.StatusOK, envelope{"status": "ok"}, nil)
}

func (app *application) handleWS(writer http.ResponseWriter, request *http.Request) {
	id, err := app.readIdParam(request)
	if err != nil {
		app.writeJSON(writer, http.StatusBadRequest, envelope{"error": err.Error()}, nil)
		return
	}
	app.logger.Info(id)
	app.writeJSON(writer, http.StatusOK, envelope{"id": id}, nil)
	return
}

func (app *application) tablesHandler(writer http.ResponseWriter, request *http.Request) {
	var message struct {
		Name    string
		Payload any
	}

	err := app.readJSON(writer, request, &message)
	if err != nil {
		app.writeJSON(writer, http.StatusBadRequest, envelope{"error": err.Error()}, nil)
		return
	}
	return
}
