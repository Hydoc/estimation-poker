package internal

import (
	"testing"

	"github.com/google/uuid"

	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		name         string
		msg          *WebsocketMessage
		expectedType string
		expectedData any
	}{
		{
			name:         "newLeave",
			msg:          newLeave("Test"),
			expectedType: leave,
			expectedData: "Test",
		},
		{
			name:         "newRoomLocked",
			msg:          newRoomLocked(),
			expectedType: roomLocked,
			expectedData: nil,
		},
		{
			name:         "newRoomOpened",
			msg:          newRoomOpened(),
			expectedType: roomOpened,
			expectedData: nil,
		},
		{
			name:         "newEveryoneIsDone",
			msg:          newEveryoneIsDone(),
			expectedType: everyoneDone,
			expectedData: nil,
		},
		{
			name:         "newNewRound",
			msg:          newNewRound(),
			expectedType: newRound,
			expectedData: nil,
		},
		{
			name:         "newReveal",
			msg:          newReveal(make(map[*Client]bool)),
			expectedType: reveal,
			expectedData: []map[string]any{},
		},
		{
			name:         "newYouSkipped",
			msg:          newYouSkipped(),
			expectedType: youSkipped,
			expectedData: nil,
		},
		{
			name:         "newYouGuessed",
			msg:          newYouGuessed(2),
			expectedType: youGuessed,
			expectedData: 2,
		},
		{
			name:         "newPermissions",
			msg:          newPermissions("Test", "Abc", uuid.New()),
			expectedType: permissions,
			expectedData: Permissions{
				CanLockRoom: false,
			},
		},
		{
			name:         "newPermissions when client and room creator have same name",
			msg:          newPermissions("Test", "Test", uuid.MustParse("67ddc335-0aa0-41f9-8289-2649da77aee7")),
			expectedType: permissions,
			expectedData: Permissions{
				CanLockRoom: true,
				Key:         uuid.MustParse("67ddc335-0aa0-41f9-8289-2649da77aee7").String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.msg.Type, tt.expectedType)
			assert.DeepEqual(t, tt.msg.Data, tt.expectedData)
		})
	}
}
