package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (srv *gameServer) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/subscribe/:roomId", srv.withRequiredQueryParam("name", srv.subscribeHandler))
	router.HandlerFunc(http.MethodPost, "/publish/:roomId", srv.publishHandler)
	return srv.recoverPanic(router)
}
