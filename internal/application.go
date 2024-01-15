package internal

import (
	"fmt"
	"github.com/Hydoc/guess-dev/internal/member"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

type Application struct {
	memberList []*member.Member
	router     *mux.Router
	upgrader   websocket.Upgrader
}

func (app *Application) ConfigureRouting() {
	app.router.HandleFunc("/room/{id}/product-owner", app.handleWs)
	app.router.HandleFunc("/room/{id}/developer", app.handleWs)
}

func (app *Application) Listen(addr string) {
	log.Fatal(http.ListenAndServe(addr, app.router))
}

func (app *Application) handleWs(writer http.ResponseWriter, request *http.Request) {
	routeParams := mux.Vars(request)
	roomId, ok := routeParams["id"]
	if !ok {
		log.Println("id is missing in parameters")
		return
	}
	name := request.URL.Query().Get("name")
	if len(name) == 0 {
		log.Println("name is missing in query")
		return
	}

	connection, err := app.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	var memberType string
	if strings.Contains(request.URL.Path, "product-owner") {
		memberType = member.ProductOwner
	} else {
		memberType = member.Developer
	}

	newMember := member.NewMember(name, roomId, memberType, connection)
	app.memberList = append(app.memberList, newMember)
	app.broadcastInRoom(roomId, fmt.Sprintf("%s joined.", newMember.Name))
	newMember.Reader(app.broadcastInRoom)
}

func (app *Application) broadcastInRoom(roomId, message string) {
	for _, m := range app.memberList {
		if m.RoomId == roomId {
			m.Send([]byte(message))
		}
	}
}

func NewApplication(memberList []*member.Member, router *mux.Router, upgrader websocket.Upgrader) *Application {
	return &Application{
		memberList: memberList,
		router:     router,
		upgrader:   upgrader,
	}
}
