package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (srv *gameServer) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/v1/health", srv.healthHandler)

	router.HandlerFunc(http.MethodPost, "/v1/rooms", srv.createRoomHandler)

	router.HandlerFunc(http.MethodGet, "/v1/rooms/subscribe/:roomId", srv.withRequiredQueryParam("name", srv.subscribeHandler))
	router.HandlerFunc(http.MethodPost, "/v1/rooms/publish/:roomId", srv.publishHandler)
	return srv.recoverPanic(router)
}
