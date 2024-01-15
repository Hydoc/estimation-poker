package main

import (
	"github.com/Hydoc/guess-dev/internal"
	"github.com/Hydoc/guess-dev/internal/member"
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
	app := internal.NewApplication([]*member.Member{}, mux.NewRouter(), upgrader)
	app.ConfigureRouting()
	app.Listen(":8080")
}
