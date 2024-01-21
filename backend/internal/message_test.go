package internal

import (
	"reflect"
	"testing"
)

func TestMessage_ToJson(t *testing.T) {
	messageTests := []struct {
		message     message
		expectedDTO messageDTO
	}{
		{
			message: newJoin(),
			expectedDTO: messageDTO{
				"type": "join",
			},
		},
		{
			message: newLeave(),
			expectedDTO: messageDTO{
				"type": "leave",
			},
		},
		{
			message: newResetRound(),
			expectedDTO: messageDTO{
				"type": "reset-round",
			},
		},
		{
			message: newDeveloperGuessed(),
			expectedDTO: messageDTO{
				"type": "developer-guessed",
			},
		},
		{
			message: newEveryoneGuessed(),
			expectedDTO: messageDTO{
				"type": "everyone-guessed",
			},
		},
		{
			message: newYouGuessed(2),
			expectedDTO: messageDTO{
				"type": "you-guessed",
				"data": 2,
			},
		},
		{
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

	for _, testCase := range messageTests {
		got := testCase.message.ToJson()
		want := testCase.expectedDTO
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, got %v", want, got)
		}
	}
}
