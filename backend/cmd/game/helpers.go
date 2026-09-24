package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func readIdParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(httprouter.ParamsFromContext(r.Context()).ByName("roomId"))
	if err != nil {
		return uuid.Nil, errors.New("invalid roomId param")
	}
	return id, nil
}

func readNameQueryParam(r *http.Request) (string, error) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		return "", errors.New("invalid name query parameter")
	}
	return name, nil
}

func writeTimeout(ctx context.Context, timeout time.Duration, conn *websocket.Conn, msg any) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return wsjson.Write(ctx, conn, msg)
}

func (srv *gameServer) writeJSON(writer http.ResponseWriter, status int, data any, headers http.Header) error {
	jsonResponse, err := json.Marshal(data, json.Deterministic(true))
	if err != nil {
		return err
	}

	jsonResponse = append(jsonResponse, '\n')

	maps.Copy(writer.Header(), headers)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	writer.Write(jsonResponse)

	return nil
}
