package main

import (
	"github.com/Hydoc/guess-dev/backend/internal"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	hub := internal.NewHub()
	go hub.Run()
	app := internal.NewApplication(mux.NewRouter(), upgrader, hub)
	router := app.ConfigureRouting()
	log.Fatal(http.ListenAndServe(":8080", router))
}
