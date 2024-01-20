package main

import (
	"github.com/Hydoc/guess-dev/backend/internal"
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
	app.ConfigureRouting()
	app.Listen(":8080")
}
