package main

import (
	"encoding/json/v2"
	"maps"
	"net/http"
)

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
