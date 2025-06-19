package game

import "github.com/coder/websocket"

type Player struct {
	name string
	room *Room
	conn *websocket.Conn
}

func NewPlayer(name string, room *Room, conn *websocket.Conn) *Player {
	return &Player{
		name: name,
		room: room,
		conn: conn,
	}
}
