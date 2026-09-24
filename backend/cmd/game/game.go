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
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type gameServer struct {
	logger *slog.Logger

	publishLimiter *rate.Limiter

	roomsMu sync.RWMutex
	rooms   map[uuid.UUID]*room

	handlerRegistry *messageHandlerRegistry
}

func newGameServer(logger *slog.Logger, handlerRegistry *messageHandlerRegistry) *gameServer {
	return &gameServer{
		logger:          logger,
		publishLimiter:  rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
		rooms:           make(map[uuid.UUID]*room),
		handlerRegistry: handlerRegistry,
	}
}

func (srv *gameServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	roomId, err := readIdParam(r)
	if err != nil {
		srv.badRequestResponse(w, r, err)
		return
	}

	name, err := readNameQueryParam(r)
	if err != nil {
		srv.badRequestResponse(w, r, err)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		srv.logger.Error(err.Error())
		return
	}

	defer conn.Close(websocket.StatusInternalError, "")

	err = srv.subscribeRoom(r.Context(), conn, name, roomId)

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
		srv.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		Message incomingMessage `json:"message"`
	}

	err = json.UnmarshalRead(r.Body, &input)
	if err != nil {
		srv.badRequestResponse(w, r, err)
		return
	}

	srv.roomsMu.RLock()
	foundRoom, roomExists := srv.rooms[roomId]
	srv.roomsMu.RUnlock()

	if !roomExists {
		srv.notFoundResponse(w, r)
		return
	}

	srv.handlerRegistry.handlersMu.RLock()
	handler, handlerExists := srv.handlerRegistry.handlers[input.Message.Type]
	srv.handlerRegistry.handlersMu.RUnlock()
	if !handlerExists {
		srv.notFoundResponse(w, r)
		return
	}

	result, err := handler(foundRoom, input.Message.Data)
	if err != nil {
		srv.badRequestResponse(w, r, err)
		return
	}

	srv.publishRoom(result, roomId)
}

func (srv *gameServer) subscribeRoom(ctx context.Context, conn *websocket.Conn, name string, roomId uuid.UUID) error {
	ctx = conn.CloseRead(ctx)

	s := newSubscriber(name, conn)

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

func (srv *gameServer) publishRoom(msg outgoingMessage, roomId uuid.UUID) {
	srv.roomsMu.RLock()
	r, exists := srv.rooms[roomId]
	srv.roomsMu.RUnlock()

	if !exists {
		return
	}

	if err := srv.publishLimiter.Wait(context.Background()); err != nil {
		return
	}

	r.publish(msg)
}

func (srv *gameServer) addRoomSubscriber(s *subscriber, roomId uuid.UUID) {
	srv.roomsMu.RLock()
	r, exists := srv.rooms[roomId]
	srv.roomsMu.RUnlock()

	if !exists {
		srv.roomsMu.Lock()
		r, exists = srv.rooms[roomId]
		if !exists {
			r = newRoom()
			srv.rooms[roomId] = r
		}
		srv.roomsMu.Unlock()
	}

	r.addSubscriber(s)
}

func (srv *gameServer) deleteRoomSubscriber(s *subscriber, roomId uuid.UUID) {
	srv.roomsMu.RLock()
	r, exists := srv.rooms[roomId]
	srv.roomsMu.RUnlock()

	if !exists {
		return
	}

	r.deleteSubscriber(s)
}
