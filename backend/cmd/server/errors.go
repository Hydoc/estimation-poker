package main

import "net/http"

func (app *application) logError(request *http.Request, err error) {
	var (
		method = request.Method
		uri    = request.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (app *application) errorResponse(writer http.ResponseWriter, request *http.Request, status int, message any) {
	err := app.writeJSON(writer, status, envelope{"error": message}, nil)
	if err != nil {
		app.logError(request, err)
		writer.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(writer http.ResponseWriter, request *http.Request, err error) {
	app.logError(request, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(writer, request, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(writer http.ResponseWriter, request *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(writer, request, http.StatusNotFound, message)
}

func (app *application) badRequestResponse(writer http.ResponseWriter, request *http.Request, err error) {
	app.errorResponse(writer, request, http.StatusBadRequest, err.Error())
}
