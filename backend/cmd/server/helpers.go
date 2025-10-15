package main

import (
	"encoding/json"
	"maps"
	"net/http"
)

func (app *application) writeJson(writer http.ResponseWriter, status int, data any, headers http.Header) error {
	jsonResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// nicety for terminal apps
	jsonResponse = append(jsonResponse, '\n')

	maps.Copy(writer.Header(), headers)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	writer.Write(jsonResponse)

	return nil
}
