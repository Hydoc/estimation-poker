package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (srv *gameServer) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(srv.notFoundResponse)
	router.HandlerFunc(http.MethodGet, "/v1/health", srv.healthHandler)

	router.HandlerFunc(http.MethodPost, "/v1/rooms", srv.createRoomHandler)

	router.HandlerFunc(http.MethodGet, "/v1/rooms/subscribe/:roomId", alice.New(srv.withRequiredQueryParam("name")).ThenFunc(srv.subscribeHandler).(http.HandlerFunc))
	router.HandlerFunc(http.MethodPost, "/v1/rooms/publish/:roomId", srv.publishHandler)

	return alice.New(srv.recoverPanic).Then(router)
}
