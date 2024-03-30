package internal

type RoomId string

type Room struct {
	id         RoomId
	InProgress bool
	leave      chan *Client
	join       chan *Client
	clients    map[*Client]bool
	broadcast  chan message
	destroy    chan<- RoomId
}

func newRoom(name RoomId, destroy chan<- RoomId) *Room {
	return &Room{
		id:         name,
		InProgress: false,
		leave:      make(chan *Client),
		join:       make(chan *Client),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan message),
		destroy:    destroy,
	}
}

func (room *Room) everyDevGuessed() bool {
	for client := range room.clients {
		if client.Role == Developer && client.Guess == 0 {
			return false
		}
	}
	return true
}

func (room *Room) Run() {
	for {
		select {
		case client := <-room.join:
			room.clients[client] = true
		case client := <-room.leave:
			if _, ok := room.clients[client]; ok {
				delete(room.clients, client)
			}
			if len(room.clients) == 0 {
				room.destroy <- room.id
			}
		case msg := <-room.broadcast:
			for client := range room.clients {
				switch msg.(type) {
				case clientMessage:
					if msg.(clientMessage).isEstimate() {
						room.InProgress = true
					}
					client.send <- msg
				case developerGuessed:
					if room.everyDevGuessed() {
						client.send <- newEveryoneGuessed()
						continue
					}
					client.send <- msg
				case resetRound:
					room.InProgress = false
					if client.Role == Developer {
						client.reset()
					}
					client.send <- msg
				case leave:
					if room.InProgress {
						room.InProgress = false
						if client.Role == Developer {
							client.reset()
						}
						client.send <- newResetRound()
						continue
					}
					client.send <- msg
				default:
					client.send <- msg
				}
			}
		}
	}
}
