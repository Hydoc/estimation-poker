package main

import (
	"encoding/json/v2"
	"net/http"
	"testing"

	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func Test_gameServer_healthHandler(t *testing.T) {
	want := envelope{
		"status": "available",
		"systemInfo": map[string]any{
			"environment": "dev",
			"version":     version,
		},
	}
	app := newTestGameServer(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	response := ts.get(t, "/v1/health")

	var got envelope
	err := json.Unmarshal(response.body, &got)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Equal(t, http.StatusOK, response.status)
	assert.DeepEqual(t, got, want)
}
