package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) writeJson(writer http.ResponseWriter, status int, data any, headers http.Header) error {
	jsonResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// nicety for terminal apps
	jsonResponse = append(jsonResponse, '\n')

	for key, value := range headers {
		writer.Header()[key] = value
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	writer.Write(jsonResponse)

	return nil
}
