package internal

type roomBroadcastMessage struct {
	RoomId  string
	message message
}

type Hub struct {
	roomBroadcast chan roomBroadcastMessage
	register      chan *Client
	Unregister    chan *Client
	clients       map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		roomBroadcast: make(chan roomBroadcastMessage),
		register:      make(chan *Client),
		Unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
	}
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
		case client := <-hub.Unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
			}
		case msg := <-hub.roomBroadcast:
			for client := range hub.clients {
				if client.RoomId == msg.RoomId {
					switch msg.message.(type) {
					case developerGuessed:
						if hub.everyDevInRoomGuessed(msg.RoomId) {
							client.send <- newEveryoneGuessed()
							continue
						}
						client.send <- msg.message
					case resetRound:
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
