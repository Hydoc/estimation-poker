package main

import (
	"fmt"
	"github.com/Hydoc/guess-dev/internal/member"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var memberList []*member.Member

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func joinRoom(roomId, name, memberType string, connection *websocket.Conn) {
	newMember := member.NewMember(name, roomId, memberType, connection)
	memberList = append(memberList, newMember)
	broadcastInRoom(roomId, fmt.Sprintf("%s joined.", newMember.Name))
	newMember.Reader(broadcastInRoom)
}

func broadcastInRoom(roomId, message string) {
	for _, m := range memberList {
		if m.RoomId == roomId {
			m.Send([]byte(message))
		}
	}
}

func handleWebsocket(memberType string) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		routeParams := mux.Vars(request)
		id, ok := routeParams["id"]
		if !ok {
			log.Println("id is missing in parameters")
			return
		}
		name, ok := routeParams["name"]
		if !ok {
			log.Println("name is missing in parameters")
			return
		}

		joinRoom(id, name, memberType, connection)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/room/{id}/product-owner/{name}", handleWebsocket(member.ProductOwner))
	router.HandleFunc("/room/{id}/developer/{name}", handleWebsocket(member.Developer))
	log.Fatal(http.ListenAndServe(":8080", router))
}
