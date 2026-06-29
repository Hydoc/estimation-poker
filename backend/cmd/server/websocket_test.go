package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/Hydoc/estimation-poker/backend/internal"
	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func TestApplication_handleWs(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		rooms          map[uuid.UUID]*internal.Room
		expectedError  map[string]string
		expectedRoomId string
		expectedRole   string
		expectedStatus int
	}{
		{
			name: "connect as developer",
			url:  "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/developer?name=Test",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"): {
					Id: uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"),
				},
			},
			expectedError:  nil,
			expectedStatus: 101,
			expectedRoomId: "ffb25a3d-a5db-42b7-9733-345f61167077",
			expectedRole:   internal.Developer,
		},
		{
			name: "connect as product owner",
			url:  "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=Test",
			rooms: map[uuid.UUID]*internal.Room{
				uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"): {
					Id: uuid.MustParse("ffb25a3d-a5db-42b7-9733-345f61167077"),
				},
			},
			expectedError:  nil,
			expectedStatus: 101,
			expectedRoomId: "ffb25a3d-a5db-42b7-9733-345f61167077",
			expectedRole:   internal.ProductOwner,
		},
		{
			name:  "not connecting due to name too long",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=whateverthisisitiswaytoooooooooooolong",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"error": "name must be smaller or equal to 15",
			},
			expectedStatus: 400,
		},
		{
			name:  "not connecting due to missing name",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"message": "name is missing in query",
			},
			expectedStatus: 400,
			expectedRoomId: "ffb25a3d-a5db-42b7-9733-345f61167077",
			expectedRole:   internal.ProductOwner,
		},
		{
			name:  "not connecting due to invalid roomId not found",
			url:   "/v1/room/invalid/product-owner?name=test",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"error": "invalid id parameter",
			},
			expectedStatus: 400,
		},
		{
			name:  "not connecting because room not found",
			url:   "/v1/room/ffb25a3d-a5db-42b7-9733-345f61167077/product-owner?name=test",
			rooms: make(map[uuid.UUID]*internal.Room),
			expectedError: map[string]string{
				"error": "the requested resource could not be found",
			},
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApplication(t, tt.rooms)

			ts := newTestServer(t, app.routes())
			defer ts.Close()

			url := "ws" + strings.TrimPrefix(ts.URL, "http") + tt.url
			_, response, _ := websocket.Dial(context.Background(), url, nil)

			assert.Equal(t, response.StatusCode, tt.expectedStatus)

			if tt.expectedError != nil {
				var got map[string]string
				json.NewDecoder(response.Body).Decode(&got)
				assert.DeepEqual(t, got, tt.expectedError)
				return
			}
		})
	}
}
