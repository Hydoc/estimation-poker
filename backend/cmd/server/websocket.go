package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/coder/websocket"

	"github.com/Hydoc/estimation-poker/backend/internal"
)

func (app *application) handleWs(writer http.ResponseWriter, request *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	roomId, err := app.readIdParam(request)
	if err != nil {
		app.badRequestResponse(writer, request, err)
		return
	}

	name := request.URL.Query().Get("name")

	if utf8.RuneCountInString(name) > 15 {
		app.badRequestResponse(writer, request, errors.New("name must be smaller or equal to 15"))
		return
	}

	clientRoom, ok := app.rooms[roomId]
	if !ok {
		app.notFoundResponse(writer, request)
		return
	}

	connection, err := websocket.Accept(writer, request, nil)
	if err != nil {
		app.logger.Info(fmt.Sprintf("upgrade: %s", err))
		return
	}

	clientRole := internal.Developer
	if strings.Contains(request.URL.Path, "product-owner") {
		clientRole = internal.ProductOwner
	}
	client := internal.NewClient(name, clientRole, clientRoom, connection, app.bus, app.logger)

	go client.WebsocketReader()
	go client.WebsocketWriter()
	clientRoom.Join(client)
}
