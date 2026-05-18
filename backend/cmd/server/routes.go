package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodPost, "/v1/room", app.createNewRoom)
	router.HandlerFunc(http.MethodPost, "/v1/room/:id/connection-state", app.handleConnectionState)

	router.HandlerFunc(http.MethodGet, "/v1/room/:id/product-owner", app.withRequiredQueryParam("name", app.handleWs))
	router.HandlerFunc(http.MethodGet, "/v1/rooms", app.handleFetchActiveRooms)
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/metadata", app.handleFetchRoomMetadata)
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/developer", app.withRequiredQueryParam("name", app.handleWs))
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/state", app.handleFetchRoomState)

	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthcheckHandler)

	return app.recoverPanic(router)
}
