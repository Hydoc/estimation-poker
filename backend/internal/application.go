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
	// Query ?name=NAME required
	app.router.HandleFunc("/room/{id}/product-owner", app.handleWs)
	app.router.HandleFunc("/room/{id}/developer", app.handleWs)
	app.router.HandleFunc("/room/{id}/users/exists", app.handleUserInRoomExists)

	app.router.HandleFunc("/room/{id}/users", app.handleFetchUsers)
}

func (app *Application) Listen(addr string) {
	log.Fatal(http.ListenAndServe(addr, app.router))
}

func (app *Application) handleUserInRoomExists(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Content-Type", "application/json")
	if request.Method != http.MethodGet {
		writer.WriteHeader(405)
		return
	}

	roomId, ok := mux.Vars(request)["id"]
	if !ok {
		r := map[string]string{
			"message": "id is missing in parameters",
		}
		writer.WriteHeader(400)
		json.NewEncoder(writer).Encode(r)
		return
	}

	name := request.URL.Query().Get("name")
	if len(name) == 0 {
		writer.WriteHeader(400)
		json.NewEncoder(writer).Encode(map[string]string{
			"message": "name is missing in query",
		})
		return
	}

	for _, mem := range app.memberList {
		if mem.Name() == name && roomId == mem.RoomId() {
			writer.WriteHeader(409)
			json.NewEncoder(writer).Encode(map[string]bool{
				"exists": true,
			})
			return
		}
	}
	json.NewEncoder(writer).Encode(map[string]bool{
		"exists": false,
	})
}

func (app *Application) handleFetchUsers(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Content-Type", "application/json")
	if request.Method != http.MethodGet {
		writer.WriteHeader(405)
		return
	}

	roomId, ok := mux.Vars(request)["id"]
	if !ok {
		writer.WriteHeader(400)
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
	err := json.NewEncoder(writer).Encode(usersInRoom)
	if err != nil {
		log.Println("error while encoding usersInRoom:", err)
		return
	}
}

func (app *Application) handleWs(writer http.ResponseWriter, request *http.Request) {
	routeParams := mux.Vars(request)
	roomId, ok := routeParams["id"]
	if !ok {
		writer.WriteHeader(400)
		json.NewEncoder(writer).Encode(map[string]string{
			"message": "id is missing in parameters",
		})
		return
	}
	name := request.URL.Query().Get("name")
	if len(name) == 0 {
		writer.WriteHeader(400)
		json.NewEncoder(writer).Encode(map[string]string{
			"message": "id is missing in query",
		})
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
	encodedMessage, err := json.Marshal(member.NewJoin().ToJson())
	app.broadcastInRoom(roomId, encodedMessage)
	broadcastChannel := make(chan member.Message)
	go newMember.WebsocketReader(broadcastChannel)
	app.handleBroadcastMessage(<-broadcastChannel, roomId)
}

func (app *Application) handleBroadcastMessage(broadcastMessage member.Message, roomId string) {
	switch broadcastMessage.(type) {
	case member.Leave:
		memberToRemove := broadcastMessage.(member.Leave).Payload()
		app.removeMember(memberToRemove)
		app.broadcastInRoom(roomId, app.encodeMessage(broadcastMessage))
		break
	case member.DeveloperGuessed:
		if app.everyDeveloperInRoomGuessed(roomId) {
			app.broadcastInRoom(roomId, app.encodeMessage(member.NewEveryoneGuessed()))
			return
		}
		app.broadcastInRoom(roomId, app.encodeMessage(broadcastMessage))
	default:
		app.broadcastInRoom(roomId, app.encodeMessage(broadcastMessage))
		return
	}
}

func (app *Application) broadcastInRoom(roomId string, message []byte) {
	for _, m := range app.memberList {
		if m.RoomId() == roomId {
			m.Send(message)
		}
	}
}

func (app *Application) removeMember(mem member.Member) {
	for i, m := range app.memberList {
		if m.Name() == mem.Name() && m.RoomId() == mem.RoomId() {
			app.memberList = append(app.memberList[:i], app.memberList[i+1:]...)
			break
		}
	}
}

func (app *Application) everyDeveloperInRoomGuessed(roomId string) bool {
	for _, mem := range app.memberList {
		if mem.RoomId() != roomId {
			continue
		}
		switch mem.(type) {
		case *member.Developer:
			if mem.(*member.Developer).Guess == 0 {
				return false
			}
		}
	}

	return true
}

func (app *Application) encodeMessage(message member.Message) []byte {
	encoded, err := json.Marshal(message.ToJson())
	if err != nil {
		log.Fatal("failed encoding message:", err)
		return nil
	}
	return encoded
}

func NewApplication(memberList []member.Member, router *mux.Router, upgrader websocket.Upgrader) *Application {
	return &Application{
		memberList: memberList,
		router:     router,
		upgrader:   upgrader,
	}
}
