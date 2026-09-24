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
	roomId, err := readIdParam(r)
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

	err = srv.subscribeRoom(r.Context(), conn, roomId)

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
	roomId, err := readIdParam(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var input struct {
		Message incomingMessage `json:"message"`
	}

	err = json.UnmarshalRead(r.Body, &input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	cmd, err := fabricateMessage(input.Message)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	srv.publishRoom(cmd, roomId)
}

func (srv *gameServer) subscribeRoom(ctx context.Context, conn *websocket.Conn, roomId uuid.UUID) error {
	ctx = conn.CloseRead(ctx)

	s := &subscriber{
		messages: make(chan any),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up")
		},
	}

	srv.addRoomSubscriber(s, roomId)
	defer srv.deleteRoomSubscriber(s, roomId)

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

func (srv *gameServer) publishRoom(msg message, roomId uuid.UUID) {
	srv.roomsMu.Lock()
	defer srv.roomsMu.Unlock()

	srv.publishLimiter.Wait(context.Background())

	for s := range srv.rooms[roomId] {
		select {
		case s.messages <- msg.ToBroadcastPayload():
		default:
			go s.closeSlow()
		}
	}
}

func (srv *gameServer) addRoomSubscriber(s *subscriber, roomId uuid.UUID) {
	srv.roomsMu.Lock()
	if len(srv.rooms[roomId]) == 0 {
		srv.rooms[roomId] = make(map[*subscriber]struct{})
	}
	srv.rooms[roomId][s] = struct{}{}
	srv.roomsMu.Unlock()
}

func (srv *gameServer) deleteRoomSubscriber(s *subscriber, roomId uuid.UUID) {
	srv.roomsMu.Lock()
	if len(srv.rooms[roomId]) != 0 {
		delete(srv.rooms[roomId], s)
	}
	srv.roomsMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, conn *websocket.Conn, msg any) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return wsjson.Write(ctx, conn, msg)
}

type subscriber struct {
	messages  chan any
	closeSlow func()
}
