package main

import (
	"net/http"
)

type envelope map[string]any

func (srv *gameServer) logError(request *http.Request, err error) {
	var (
		method = request.Method
		uri    = request.URL.RequestURI()
	)
	srv.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (srv *gameServer) errorResponse(writer http.ResponseWriter, request *http.Request, status int, message any) {
	err := srv.writeJSON(writer, status, envelope{"error": message}, nil)
	if err != nil {
		srv.logError(request, err)
		writer.WriteHeader(500)
	}
}

func (srv *gameServer) serverErrorResponse(writer http.ResponseWriter, request *http.Request, err error) {
	srv.logError(request, err)
	msg := "the server encountered a problem and could not process your request"
	srv.errorResponse(writer, request, http.StatusInternalServerError, msg)
}

func (srv *gameServer) notFoundResponse(writer http.ResponseWriter, request *http.Request) {
	msg := "the requested resource could not be found"
	srv.errorResponse(writer, request, http.StatusNotFound, msg)
}

func (srv *gameServer) badRequestResponse(writer http.ResponseWriter, request *http.Request, err error) {
	srv.errorResponse(writer, request, http.StatusBadRequest, err.Error())
}
