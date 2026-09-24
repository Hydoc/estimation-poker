package internal

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"golang.org/x/time/rate"
)

type gameServer struct {
	logger *slog.Logger

	serveMux http.ServeMux

	publishLimiter *rate.Limiter

	roomsMu sync.Mutex
	rooms   map[string]map[*subscriber]struct{}
}

func newGameServer() *gameServer {
	srv := &gameServer{
		rooms: make(map[string]map[*subscriber]struct{}),
	}

	srv.serveMux.HandleFunc("GET /subscribe", srv.subscribeHandler)
	srv.serveMux.HandleFunc("POST /publish", srv.publishHandler)

	return srv
}

func (srv *gameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.serveMux.ServeHTTP(w, r)
}

func (srv *gameServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		srv.logger.Error(err.Error())
		return
	}

	defer conn.Close(websocket.StatusInternalError, "")
	room := strings.Split(r.URL.Path, "/")

	if len(room) != 3 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	srv.logger.Info("room", "room", room[2])
	err = srv.subscribeRoom(r.Context(), conn, room[2])

	if errors.Is(err, context.Canceled) {
		return
	}

	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		srv.logger.Error(err.Error())
		return
	}
}

func (srv *gameServer) publishHandler(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, 8192)
	defer body.Close()
	msg, err := io.ReadAll(body)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	room := strings.Split(r.URL.Path, "/")
	if len(room) != 3 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	srv.publishRoom(msg, room[2])
}

func (srv *gameServer) subscribeRoom(ctx context.Context, conn *websocket.Conn, room string) error {
	ctx = conn.CloseRead(ctx)

	s := &subscriber{
		msgs: make(chan []byte),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up")
		},
	}

	srv.addRoomSubscriber(s, room)
	defer srv.deleteRoomSubscriber(s, room)

	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*5, conn, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (srv *gameServer) publishRoom(msg []byte, room string) {
	srv.roomsMu.Lock()
	defer srv.roomsMu.Unlock()

	srv.publishLimiter.Wait(context.Background())

	for s := range srv.rooms[room] {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}

func (srv *gameServer) addRoomSubscriber(s *subscriber, room string) {
	srv.roomsMu.Lock()
	if len(srv.rooms[room]) == 0 {
		srv.rooms[room] = make(map[*subscriber]struct{})
	}
	srv.rooms[room][s] = struct{}{}
	srv.roomsMu.Unlock()
}

func (srv *gameServer) deleteRoomSubscriber(s *subscriber, room string) {
	srv.roomsMu.Lock()
	if len(srv.rooms[room]) != 0 {
		delete(srv.rooms[room], s)
	}
	srv.roomsMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, conn *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return conn.Write(ctx, websocket.MessageText, msg)
}

type subscriber struct {
	msgs      chan []byte
	closeSlow func()
}
