package main

import "net/http"

func (app *application) Routes() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /api/estimation/room/{id}/product-owner", app.withRequiredQueryParam("name", app.handleWs))
	router.HandleFunc("GET /api/estimation/room/{id}/developer", app.withRequiredQueryParam("name", app.handleWs))
	router.HandleFunc("GET /api/estimation/room/{id}/users/exists", app.withRequiredQueryParam("name", app.handleUserInRoomExists))
	router.HandleFunc("GET /api/estimation/room/{id}/users", app.handleFetchUsers)
	router.HandleFunc("GET /api/estimation/room/{id}/{username}/permissions", app.handleFetchPermissions)
	router.HandleFunc("GET /api/estimation/room/{id}/state", app.handleFetchRoomState)
	router.HandleFunc("GET /api/estimation/room/rooms", app.handleFetchActiveRooms)
	router.HandleFunc("GET /api/estimation/possible-guesses", app.handlePossibleGuesses)
	router.HandleFunc("POST /api/estimation/room/{id}/authenticate", app.handleRoomAuthenticate)
	return router
}
