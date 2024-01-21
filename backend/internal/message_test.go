package internal

import (
	"reflect"
	"testing"
)

func TestMessage_ToJson(t *testing.T) {
	testSuites := []struct {
		name        string
		message     message
		expectedDTO messageDTO
	}{
		{
			name:    "join",
			message: newJoin(),
			expectedDTO: messageDTO{
				"type": "join",
			},
		},
		{
			name:    "leave",
			message: newLeave(),
			expectedDTO: messageDTO{
				"type": "leave",
			},
		},
		{
			name:    "reset round",
			message: newResetRound(),
			expectedDTO: messageDTO{
				"type": "reset-round",
			},
		},
		{
			name:    "developer guessed",
			message: newDeveloperGuessed(),
			expectedDTO: messageDTO{
				"type": "developer-guessed",
			},
		},
		{
			name:    "everyone guessed",
			message: newEveryoneGuessed(),
			expectedDTO: messageDTO{
				"type": "everyone-guessed",
			},
		},
		{
			name:    "you guessed",
			message: newYouGuessed(2),
			expectedDTO: messageDTO{
				"type": "you-guessed",
				"data": 2,
			},
		},
		{
			name: "any client message",
			message: clientMessage{
				Type: "test",
				Data: "any",
			},
			expectedDTO: messageDTO{
				"type": "test",
				"data": "any",
			},
		},
	}

	for _, suite := range testSuites {
		t.Run(suite.name, func(t *testing.T) {
			got := suite.message.ToJson()
			want := suite.expectedDTO
			if !reflect.DeepEqual(got, want) {
				t.Errorf("expected %v, got %v", want, got)
			}
		})
	}
}
