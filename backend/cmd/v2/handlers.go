package main

import "net/http"

func (app *application) healthcheckHandler(writer http.ResponseWriter, request *http.Request) {
	app.writeJSON(writer, http.StatusOK, envelope{"status": "ok"}, nil)
}
