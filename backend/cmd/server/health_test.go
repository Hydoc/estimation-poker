package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/Hydoc/estimation-poker/backend/internal"
	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func TestApplication_healthcheckHandler(t *testing.T) {
	want := envelope{
		"status": "available",
		"systemInfo": map[string]any{
			"environment": "dev",
			"version":     version,
		},
	}
	app := newTestApplication(t, make(map[uuid.UUID]*internal.Room))
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	response := ts.get(t, "/v1/health")

	var got envelope
	json.Unmarshal(response.body, &got)

	assert.Equal(t, http.StatusOK, response.status)
	assert.DeepEqual(t, got, want)
}
