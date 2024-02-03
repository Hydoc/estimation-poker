package internal

import (
	"reflect"
	"testing"
)

func TestMessage_ToJson(t *testing.T) {
	tests := []struct {
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.message.ToJson()
			want := test.expectedDTO
			if !reflect.DeepEqual(got, want) {
				t.Errorf("expected %v, got %v", want, got)
			}
		})
	}
}

func TestClientMessage_isEstimate(t *testing.T) {
	tests := []struct {
		name    string
		message clientMessage
		want    bool
	}{
		{
			name: "is estimate",
			want: true,
			message: clientMessage{
				Type: estimate,
				Data: nil,
			},
		},
		{
			name: "is not estimate",
			want: false,
			message: clientMessage{
				Type: "any other",
				Data: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.message.isEstimate()
			if got != test.want {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}
