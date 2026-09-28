package main

import (
	"net/http"
)

type envelope map[string]any

func (srv *gameServer) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	srv.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (srv *gameServer) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	err := srv.writeJSON(w, status, envelope{"error": message}, nil)
	if err != nil {
		srv.logError(r, err)
		w.WriteHeader(500)
	}
}

func (srv *gameServer) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	srv.logError(r, err)
	msg := "the server encountered a problem and could not process your request"
	srv.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (srv *gameServer) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "the requested resource could not be found"
	srv.errorResponse(w, r, http.StatusNotFound, msg)
}

func (srv *gameServer) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	srv.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (srv *gameServer) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	srv.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
