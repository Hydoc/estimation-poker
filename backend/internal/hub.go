package internal

type roomBroadcastMessage struct {
	RoomId  string
	message message
}

type Hub struct {
	roomBroadcast chan roomBroadcastMessage
	register      chan *Client
	unregister    chan *Client
	clients       map[*Client]bool
	rooms         map[string]bool
}

func NewHub() *Hub {
	return &Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]bool),
	}
}

func (hub *Hub) IsRoundInRoomInProgress(roomId string) bool {
	inProgress, ok := hub.rooms[roomId]
	if !ok {
		return false
	}
	return inProgress
}

func newRoomBroadcast(roomId string, message message) roomBroadcastMessage {
	return roomBroadcastMessage{
		RoomId:  roomId,
		message: message,
	}
}

func (hub *Hub) everyDevInRoomGuessed(roomId string) bool {
	for client := range hub.clients {
		if client.RoomId == roomId && client.Role == Developer && client.Guess == 0 {
			return false
		}
	}
	return true
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client] = true
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
			}
		case msg := <-hub.roomBroadcast:
			for client := range hub.clients {
				if client.RoomId == msg.RoomId {
					switch msg.message.(type) {
					case clientMessage:
						if msg.message.(clientMessage).Type == "estimate" {
							hub.rooms[msg.RoomId] = true
						}
						client.send <- msg.message
					case developerGuessed:
						if hub.everyDevInRoomGuessed(msg.RoomId) {
							client.send <- newEveryoneGuessed()
							continue
						}
						client.send <- msg.message
					case resetRound:
						if _, ok := hub.rooms[msg.RoomId]; ok {
							delete(hub.rooms, msg.RoomId)
						}
						if client.Role == Developer {
							client.reset()
						}
						client.send <- msg.message
					default:
						client.send <- msg.message
					}
				}
			}
		}
	}
}
