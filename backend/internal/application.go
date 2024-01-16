package internal

import (
	"encoding/json"
	"github.com/Hydoc/guess-dev/backend/internal/member"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

type Application struct {
	memberList []member.Member
	router     *mux.Router
	upgrader   websocket.Upgrader
}

func (app *Application) ConfigureRouting() {
	app.router.HandleFunc("/room/{id}/product-owner", app.handleWs)
	app.router.HandleFunc("/room/{id}/developer", app.handleWs)
	app.router.HandleFunc("/room/{id}/users", app.handleFetchUsers)
}

func (app *Application) Listen(addr string) {
	log.Fatal(http.ListenAndServe(addr, app.router))
}

func (app *Application) handleFetchUsers(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Content-Type", "application/json")

	roomId, ok := mux.Vars(request)["id"]
	if !ok {
		log.Println("users: id is missing in parameters")
		r := map[string]string{
			"message": "id is missing in parameters",
		}
		json.NewEncoder(writer).Encode(r)
		return
	}

	var usersInRoom = map[string][]member.UserDTO{
		"productOwnerList": {},
		"developerList":    {},
	}

	for _, mem := range app.memberList {
		switch mem.(type) {
		case *member.Developer:
			if mem.RoomId() == roomId {
				usersInRoom["developerList"] = append(usersInRoom["developerList"], mem.ToJson())
			}
			break
		case *member.ProductOwner:
			if mem.RoomId() == roomId {
				usersInRoom["productOwnerList"] = append(usersInRoom["productOwnerList"], mem.ToJson())
			}
			break
		default:
			break
		}
	}
	json.NewEncoder(writer).Encode(usersInRoom)
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

	var newMember member.Member
	if strings.Contains(request.URL.Path, "product-owner") {
		newMember = member.NewProductOwner(name, roomId, connection)
	} else {
		newMember = member.NewDeveloper(name, roomId, connection)
	}

	app.memberList = append(app.memberList, newMember)
	app.broadcastInRoom(roomId, "join")
	broadcastChannel := make(chan interface{})
	go newMember.WebsocketReader(broadcastChannel)
	app.handleBroadcastMessage(<-broadcastChannel, roomId)
}

func (app *Application) handleBroadcastMessage(broadcastMessage interface{}, roomId string) {
	switch broadcastMessage.(type) {
	case member.Leave:
		app.removeMember(broadcastMessage.(member.Leave).Payload())
		app.broadcastInRoom(roomId, "leave")
		break
	default:
		return
	}
}

func (app *Application) broadcastInRoom(roomId, message string) {
	for _, m := range app.memberList {
		if m.RoomId() == roomId {
			m.Send([]byte(message))
		}
	}
}

func (app *Application) removeMember(mem member.Member) {
	for i, m := range app.memberList {
		if m.Name() == mem.Name() && m.RoomId() == mem.RoomId() {
			app.memberList = append(app.memberList[:i], app.memberList[i+1:]...)
		}
	}
}

func NewApplication(memberList []member.Member, router *mux.Router, upgrader websocket.Upgrader) *Application {
	return &Application{
		memberList: memberList,
		router:     router,
		upgrader:   upgrader,
	}
}
