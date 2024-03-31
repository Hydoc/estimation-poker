package internal

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
)

type Application struct {
	router      *http.ServeMux
	upgrader    *websocket.Upgrader
	guessConfig *GuessConfig
	rooms       map[RoomId]*Room
	destroyRoom chan RoomId
}

func (app *Application) ConfigureRouting() *http.ServeMux {
	app.router.HandleFunc("GET /api/estimation/room/{id}/product-owner", app.withRequiredQueryParam("name", app.handleWs))
	app.router.HandleFunc("GET /api/estimation/room/{id}/developer", app.withRequiredQueryParam("name", app.handleWs))
	app.router.HandleFunc("GET /api/estimation/room/{id}/users/exists", app.contentTypeJsonMiddleware(app.withRequiredQueryParam("name", app.handleUserInRoomExists)))
	app.router.HandleFunc("GET /api/estimation/room/{id}/users", app.contentTypeJsonMiddleware(app.handleFetchUsers))
	app.router.HandleFunc("GET /api/estimation/room/{id}/{username}/permissions", app.contentTypeJsonMiddleware(app.handleFetchPermissions))
	app.router.HandleFunc("GET /api/estimation/room/{id}/state", app.contentTypeJsonMiddleware(app.handleFetchRoomState))
	app.router.HandleFunc("GET /api/estimation/room/rooms", app.contentTypeJsonMiddleware(app.handleFetchActiveRooms))
	app.router.HandleFunc("GET /api/estimation/possible-guesses", app.contentTypeJsonMiddleware(app.handlePossibleGuesses))
	app.router.HandleFunc("POST /api/estimation/room/{id}/authenticate", app.contentTypeJsonMiddleware(app.handleRoomAuthenticate))
	return app.router
}

func (app *Application) withRequiredQueryParam(param string, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		queryParam := request.URL.Query().Get(param)

		if len(queryParam) == 0 || !request.URL.Query().Has(param) {
			writer.WriteHeader(400)
			json.NewEncoder(writer).Encode(map[string]string{
				"message": fmt.Sprintf("%s is missing in query", param),
			})
			return
		}

		next.ServeHTTP(writer, request)
	}
}

func (app *Application) handleRoomAuthenticate(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	roomId := request.PathValue("id")
	actualRoom, ok := app.rooms[RoomId(roomId)]
	if !ok {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	var input struct {
		Password string `json:"password"`
	}

	err := json.NewDecoder(request.Body).Decode(&input)
	if err != nil {
		json.NewEncoder(writer).Encode(map[string]bool{
			"ok": false,
		})
		return
	}

	if actualRoom.verify(input.Password) {
		json.NewEncoder(writer).Encode(map[string]bool{
			"ok": true,
		})
		return
	}

	json.NewEncoder(writer).Encode(map[string]bool{
		"ok": false,
	})
}

func (app *Application) handleFetchPermissions(writer http.ResponseWriter, request *http.Request) {
	roomId := request.PathValue("id")
	username := request.PathValue("username")
	actualRoom, ok := app.rooms[RoomId(roomId)]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		json.NewEncoder(writer).Encode(map[string]string{
			"message": "room does not exist",
		})
		return
	}

	if actualRoom.nameOfCreator == username {
		json.NewEncoder(writer).Encode(map[string]map[string]map[string]any{
			"permissions": {
				"room": {
					"canLock": true,
					"key":     actualRoom.key.String(),
				},
			},
		})
		return
	}

	json.NewEncoder(writer).Encode(map[string]map[string]map[string]any{
		"permissions": {
			"room": {
				"canLock": false,
			},
		},
	})
}

func (app *Application) handlePossibleGuesses(writer http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(writer).Encode(app.guessConfig.Guesses)
}

func (app *Application) contentTypeJsonMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(writer, request)
	}
}

func (app *Application) handleFetchRoomState(writer http.ResponseWriter, request *http.Request) {
	roomId := request.PathValue("id")
	actualRoom, ok := app.rooms[RoomId(roomId)]
	if !ok {
		json.NewEncoder(writer).Encode(map[string]bool{
			"inProgress": false,
			"isLocked":   false,
		})
		return
	}
	json.NewEncoder(writer).Encode(map[string]bool{
		"inProgress": actualRoom.inProgress,
		"isLocked":   actualRoom.isLocked,
	})
}

func (app *Application) handleUserInRoomExists(writer http.ResponseWriter, request *http.Request) {
	roomId := request.PathValue("id")

	name := request.URL.Query().Get("name")

	if _, ok := app.rooms[RoomId(roomId)]; !ok {
		json.NewEncoder(writer).Encode(map[string]bool{
			"exists": false,
		})
		return
	}

	for client := range app.rooms[RoomId(roomId)].clients {
		if client.Name == name {
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

func (app *Application) handleFetchActiveRooms(writer http.ResponseWriter, _ *http.Request) {
	activeRooms := []string{}
	for _, room := range app.rooms {
		if !room.isLocked {
			activeRooms = append(activeRooms, string(room.id))
		}
	}
	slices.Sort(activeRooms)
	json.NewEncoder(writer).Encode(activeRooms)
}

func (app *Application) handleFetchUsers(writer http.ResponseWriter, request *http.Request) {
	roomId := request.PathValue("id")

	var usersInRoom = map[string][]userDTO{
		"productOwnerList": {},
		"developerList":    {},
	}
	var clients []*Client

	if _, ok := app.rooms[RoomId(roomId)]; !ok {
		json.NewEncoder(writer).Encode(usersInRoom)
		return
	}

	for client := range app.rooms[RoomId(roomId)].clients {
		clients = append(clients, client)
	}
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].Name < clients[j].Name
	})

	for _, c := range clients {
		switch c.Role {
		case Developer:
			usersInRoom["developerList"] = append(usersInRoom["developerList"], c.toJson())
		case ProductOwner:
			usersInRoom["productOwnerList"] = append(usersInRoom["productOwnerList"], c.toJson())
		}
	}
	json.NewEncoder(writer).Encode(usersInRoom)
}

func (app *Application) handleWs(writer http.ResponseWriter, request *http.Request) {
	roomId := request.PathValue("id")

	name := request.URL.Query().Get("name")

	connection, err := app.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	var clientRoom *Room
	if room, ok := app.rooms[RoomId(roomId)]; ok {
		// when there is a room, it's already running in a goroutine
		clientRoom = room
	} else {
		clientRoom = newRoom(RoomId(roomId), app.destroyRoom, name)
		app.rooms[clientRoom.id] = clientRoom
		go clientRoom.Run()
	}

	clientRole := Developer
	if strings.Contains(request.URL.Path, "product-owner") {
		clientRole = ProductOwner
	}
	client := newClient(name, clientRole, clientRoom, connection)
	clientRoom.join <- client
	clientRoom.broadcast <- newJoin()

	go client.websocketReader()
	go client.websocketWriter()
}

func (app *Application) ListenForRoomDestroy() {
	for {
		select {
		case roomId := <-app.destroyRoom:
			if _, ok := app.rooms[roomId]; ok {
				delete(app.rooms, roomId)
			}
		}
	}
}

func NewApplication(router *http.ServeMux, upgrader *websocket.Upgrader, config *GuessConfig) *Application {
	return &Application{
		router:      router,
		upgrader:    upgrader,
		guessConfig: config,
		rooms:       make(map[RoomId]*Room),
		destroyRoom: make(chan RoomId),
	}
}
