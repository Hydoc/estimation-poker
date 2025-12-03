package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodPost, "/api/estimation/room", app.createNewRoom)
	router.HandlerFunc(http.MethodPost, "/api/estimation/room/:id/authenticate", app.handleRoomAuthenticate)

	router.HandlerFunc(http.MethodGet, "/api/estimation/room/:id/product-owner", app.withRequiredQueryParam("name", app.handleWs))
	router.HandlerFunc(http.MethodGet, "/api/estimation/rooms", app.handleFetchActiveRooms)
	router.HandlerFunc(http.MethodGet, "/api/estimation/room/:id/developer", app.withRequiredQueryParam("name", app.handleWs))
	router.HandlerFunc(http.MethodGet, "/api/estimation/room/:id/users/exists", app.withRequiredQueryParam("name", app.handleUserInRoomExists))
	router.HandlerFunc(http.MethodGet, "/api/estimation/room/:id/users", app.handleFetchUsers)
	router.HandlerFunc(http.MethodGet, "/api/estimation/room/:id/state", app.handleFetchRoomState)
	router.HandlerFunc(http.MethodGet, "/api/estimation/room/:id/exists", app.handleRoomExists)
	router.HandlerFunc(http.MethodGet, "/api/estimation/room/:id/permissions", app.withRequiredQueryParam("name", app.handleFetchPermissions))
	router.HandlerFunc(http.MethodGet, "/api/estimation/possible-guesses", app.handlePossibleGuesses)

	// router.HandleFunc("GET /api/estimation/room/{id}/product-owner", app.withRequiredQueryParam("name", app.handleWs))
	// router.HandleFunc("GET /api/estimation/room/{id}/developer", app.withRequiredQueryParam("name", app.handleWs))
	// router.HandleFunc("GET /api/estimation/room/{id}/users/exists", app.withRequiredQueryParam("name", app.handleUserInRoomExists))
	// router.HandleFunc("GET /api/estimation/room/{id}/users", app.handleFetchUsers)
	// router.HandleFunc("GET /api/estimation/room/{id}/{username}/permissions", app.handleFetchPermissions)
	// router.HandleFunc("GET /api/estimation/room/{id}/state", app.handleFetchRoomState)
	// router.HandleFunc("GET /api/estimation/room/{id}/exists", app.handleRoomExists)
	// router.HandleFunc("GET /api/estimation/room/rooms", app.handleFetchActiveRooms)
	// router.HandleFunc("GET /api/estimation/possible-guesses", app.handlePossibleGuesses)
	// router.HandleFunc("POST /api/estimation/room/{id}/authenticate", app.handleRoomAuthenticate)
	return router
}
