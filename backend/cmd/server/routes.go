package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodPost, "/v1/room", app.createNewRoom)
	router.HandlerFunc(http.MethodPost, "/v1/room/:id/authenticate", app.handleRoomAuthenticate)

	router.HandlerFunc(http.MethodGet, "/v1/room/:id/product-owner", app.withRequiredQueryParam("name", app.handleWs))
	router.HandlerFunc(http.MethodGet, "/v1/rooms", app.handleFetchActiveRooms)
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/developer", app.withRequiredQueryParam("name", app.handleWs))
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/users/exists", app.withRequiredQueryParam("name", app.handleUserInRoomExists))
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/users", app.handleFetchUsers)
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/state", app.handleFetchRoomState)
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/exists", app.handleRoomExists)
	router.HandlerFunc(http.MethodGet, "/v1/room/:id/permissions", app.withRequiredQueryParam("name", app.handleFetchPermissions))
	router.HandlerFunc(http.MethodGet, "/v1/possible-guesses", app.handlePossibleGuesses)

	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthcheckHandler)

	return app.recoverPanic(router)
}
