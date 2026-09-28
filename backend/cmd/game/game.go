package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Hydoc/estimation-poker/backend/internal/validator"
	"github.com/coder/websocket"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

var (
	errRoomNotExists = errors.New("room does not exist")
)

type gameServer struct {
	logger *slog.Logger
	config *config

	publishLimiter *rate.Limiter

	roomsMu sync.RWMutex
	rooms   map[uuid.UUID]*room

	handlerRegistry *messageHandlerRegistry
}

func newGameServer(logger *slog.Logger, config *config, handlerRegistry *messageHandlerRegistry) *gameServer {
	return &gameServer{
		logger:          logger,
		config:          config,
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

	_, roomExists := srv.room(roomId)

	if !roomExists {
		srv.notFoundResponse(w, r)
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

func (srv *gameServer) createRoomHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RoomName string `json:"roomName"`
		Deck     string `json:"deck"`
	}

	err := json.UnmarshalRead(r.Body, &input)
	if err != nil {
		srv.badRequestResponse(w, r, err)
		return
	}

	id := uuid.New()
	createdRoom := newRoom(input.RoomName, input.Deck)

	v := validator.New()

	if validateRoom(v, createdRoom); !v.Valid() {
		srv.failedValidationResponse(w, r, v.Errors)
		return
	}

	srv.roomsMu.Lock()
	srv.rooms[id] = createdRoom
	srv.roomsMu.Unlock()

	err = srv.writeJSON(w, http.StatusOK, envelope{"id": id.String()}, nil)
	if err != nil {
		srv.serverErrorResponse(w, r, err)
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

	foundRoom, exists := srv.room(roomId)

	if !exists {
		srv.notFoundResponse(w, r)
		return
	}

	srv.handlerRegistry.handlersMu.RLock()
	handler, handlerExists := srv.handlerRegistry.handlers[input.Message.Type]
	srv.handlerRegistry.handlersMu.RUnlock()
	if !handlerExists {
		srv.serverErrorResponse(w, r, fmt.Errorf("no handler exists for the given message '%s'", input.Message.Type))
		return
	}

	result, err := handler(foundRoom, input.Message.Data)
	if err != nil {
		srv.badRequestResponse(w, r, err)
		return
	}

	srv.publishRoom(result, roomId)
}

func (srv *gameServer) room(roomId uuid.UUID) (*room, bool) {
	srv.roomsMu.RLock()
	defer srv.roomsMu.RUnlock()
	r, ok := srv.rooms[roomId]
	return r, ok
}

func (srv *gameServer) subscribeRoom(ctx context.Context, conn *websocket.Conn, name string, roomId uuid.UUID) error {
	ctx = conn.CloseRead(ctx)

	s := newSubscriber(name, conn)

	err := srv.addRoomSubscriber(s, roomId)
	if err != nil {
		return err
	}
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
	r, exists := srv.room(roomId)

	if !exists {
		return
	}

	if err := srv.publishLimiter.Wait(context.Background()); err != nil {
		return
	}

	r.publish(msg)
}

func (srv *gameServer) addRoomSubscriber(s *subscriber, roomId uuid.UUID) error {
	srv.roomsMu.RLock()
	defer srv.roomsMu.RUnlock()

	r, exists := srv.rooms[roomId]
	if !exists {
		return errRoomNotExists
	}

	r.addSubscriber(s)
	return nil
}

func (srv *gameServer) deleteRoomSubscriber(s *subscriber, roomId uuid.UUID) {
	srv.roomsMu.Lock()
	defer srv.roomsMu.Unlock()

	foundRoom, exists := srv.rooms[roomId]
	if !exists {
		return
	}

	if foundRoom.deleteSubscriber(s) {
		delete(srv.rooms, roomId)
	}
}
