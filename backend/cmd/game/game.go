package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
)

type gameServer struct {
	logger *slog.Logger

	publishLimiter *rate.Limiter

	roomsMu sync.Mutex
	rooms   map[uuid.UUID]map[*subscriber]struct{}
}

func readIdParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(httprouter.ParamsFromContext(r.Context()).ByName("roomId"))
	if err != nil {
		return uuid.Nil, errors.New("invalid roomId param")
	}
	return id, nil
}

func newGameServer(logger *slog.Logger) *gameServer {
	return &gameServer{
		logger:         logger,
		publishLimiter: rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
		rooms:          make(map[uuid.UUID]map[*subscriber]struct{}),
	}
}

func (srv *gameServer) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/subscribe/:roomId", srv.subscribeHandler)
	router.HandlerFunc(http.MethodPost, "/publish/:roomId", srv.publishHandler)
	return router
}

func (srv *gameServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIdParam(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		srv.logger.Error(err.Error())
		return
	}

	defer conn.Close(websocket.StatusInternalError, "")

	err = srv.subscribeRoom(r.Context(), conn, id)

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
	id, err := readIdParam(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var input struct {
		Message message `json:"message"`
	}

	err = json.UnmarshalRead(r.Body, &input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	srv.publishRoom(input.Message, id)
}

func (srv *gameServer) subscribeRoom(ctx context.Context, conn *websocket.Conn, room uuid.UUID) error {
	ctx = conn.CloseRead(ctx)

	s := &subscriber{
		messages: make(chan message),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up")
		},
	}

	srv.addRoomSubscriber(s, room)
	defer srv.deleteRoomSubscriber(s, room)

	for {
		select {
		case msg := <-s.messages:
			err := writeTimeout(ctx, time.Second*5, conn, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (srv *gameServer) publishRoom(msg message, room uuid.UUID) {
	srv.roomsMu.Lock()
	defer srv.roomsMu.Unlock()

	srv.publishLimiter.Wait(context.Background())

	srv.logger.Info("room", "room", room)

	for s := range srv.rooms[room] {
		select {
		case s.messages <- msg:
		default:
			go s.closeSlow()
		}
	}
}

func (srv *gameServer) addRoomSubscriber(s *subscriber, room uuid.UUID) {
	srv.roomsMu.Lock()
	if len(srv.rooms[room]) == 0 {
		srv.rooms[room] = make(map[*subscriber]struct{})
	}
	srv.rooms[room][s] = struct{}{}
	srv.roomsMu.Unlock()
}

func (srv *gameServer) deleteRoomSubscriber(s *subscriber, room uuid.UUID) {
	srv.roomsMu.Lock()
	if len(srv.rooms[room]) != 0 {
		delete(srv.rooms[room], s)
	}
	srv.roomsMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, conn *websocket.Conn, msg message) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return wsjson.Write(ctx, conn, msg)
}

type subscriber struct {
	messages  chan message
	closeSlow func()
}
