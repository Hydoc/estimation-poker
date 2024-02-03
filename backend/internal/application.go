package internal

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
)

type Application struct {
	router      *mux.Router
	upgrader    *websocket.Upgrader
	guessConfig *GuessConfig
	rooms       map[RoomId]*Room
	destroyRoom chan RoomId
}

func (app *Application) ConfigureRouting() *mux.Router {
	app.router.HandleFunc("/api/estimation/room/{id}/product-owner", app.handleWs).Queries("name", "{name:.*}")
	app.router.HandleFunc("/api/estimation/room/{id}/developer", app.handleWs).Queries("name", "{name:.*}")
	app.router.HandleFunc("/api/estimation/room/{id}/users/exists", app.handleUserInRoomExists).Methods(http.MethodGet).Queries("name", "{name:.*}")
	app.router.HandleFunc("/api/estimation/room/{id}/users", app.handleFetchUsers).Methods(http.MethodGet)
	app.router.HandleFunc("/api/estimation/room/{id}/state", app.handleRoundInRoomInProgress).Methods(http.MethodGet)
	app.router.HandleFunc("/api/estimation/room/rooms", app.handleFetchActiveRooms).Methods(http.MethodGet)
	app.router.HandleFunc("/api/estimation/possible-guesses", app.handlePossibleGuesses).Methods(http.MethodGet)
	app.router.Use(app.contentTypeJsonMiddleware)
	return app.router
}

func (app *Application) handlePossibleGuesses(writer http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(writer).Encode(app.guessConfig.Guesses)
}

func (app *Application) contentTypeJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(writer, request)
	})
}

func (app *Application) handleRoundInRoomInProgress(writer http.ResponseWriter, request *http.Request) {
	roomId := mux.Vars(request)["id"]
	if _, ok := app.rooms[RoomId(roomId)]; !ok {
		json.NewEncoder(writer).Encode(map[string]bool{
			"inProgress": false,
		})
		return
	}
	json.NewEncoder(writer).Encode(map[string]bool{
		"inProgress": app.rooms[RoomId(roomId)].InProgress,
	})
}

func (app *Application) handleUserInRoomExists(writer http.ResponseWriter, request *http.Request) {
	roomId := mux.Vars(request)["id"]

	name := request.URL.Query().Get("name")
	if len(name) == 0 {
		writer.WriteHeader(400)
		json.NewEncoder(writer).Encode(map[string]string{
			"message": "name is missing in query",
		})
		return
	}

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
		activeRooms = append(activeRooms, string(room.id))
	}
	slices.Sort(activeRooms)
	json.NewEncoder(writer).Encode(activeRooms)
}

func (app *Application) handleFetchUsers(writer http.ResponseWriter, request *http.Request) {
	roomId := mux.Vars(request)["id"]

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
	roomId := mux.Vars(request)["id"]

	name := request.URL.Query().Get("name")
	if len(name) == 0 {
		writer.WriteHeader(400)
		json.NewEncoder(writer).Encode(map[string]string{
			"message": "name is missing in query",
		})
		return
	}

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
		clientRoom = newRoom(RoomId(roomId), app.destroyRoom)
		app.rooms[clientRoom.id] = clientRoom
		go clientRoom.Run()
	}

	var client *Client
	if strings.Contains(request.URL.Path, "product-owner") {
		client = newClient(name, ProductOwner, clientRoom, connection)
	} else {
		client = newClient(name, Developer, clientRoom, connection)
	}
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

func NewApplication(router *mux.Router, upgrader *websocket.Upgrader, config *GuessConfig) *Application {
	return &Application{
		router:      router,
		upgrader:    upgrader,
		guessConfig: config,
		rooms:       make(map[RoomId]*Room),
		destroyRoom: make(chan RoomId),
	}
}
