package internal

import (
	"testing"

	"github.com/Hydoc/guess-dev/backend/internal/assert"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		name         string
		msg          *message
		expectedType string
		expectedData any
	}{
		{
			name:         "newJoin",
			msg:          newJoin(),
			expectedType: join,
			expectedData: nil,
		},
		{
			name:         "newLeave",
			msg:          newLeave(),
			expectedType: leave,
			expectedData: nil,
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
			name:         "newDeveloperGuessed",
			msg:          newDeveloperGuessed(),
			expectedType: developerGuessed,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.msg.Type, tt.expectedType)
			assert.DeepEqual(t, tt.msg.Data, tt.expectedData)
		})
	}
}
