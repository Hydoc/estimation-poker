package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/api/v2/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/api/v2/tables", app.tablesHandler)
	router.HandlerFunc(http.MethodPost, "/api/v2/tables/{id}/connect", app.handleWS)

	return router
}
