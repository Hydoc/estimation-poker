package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Hydoc/estimation-poker/backend/internal/assert"
	"github.com/Hydoc/estimation-poker/backend/internal/validator"
)

func Test_validateRoom(t *testing.T) {
	tests := []struct {
		name       string
		r          *room
		wantErrors map[string]string
	}{
		{
			name: "valid room",
			r: &room{
				name: strings.Repeat("a", 20),
				deck: "1,a2,333,E4,5",
			},
			wantErrors: map[string]string{},
		},
		{
			name: "invalid due to single number as deck",
			r: &room{
				name: "Hola",
				deck: "1",
			},
			wantErrors: map[string]string{
				"deck": fmt.Sprintf("must match %s", deckRegex.String()),
			},
		},
		{
			name: "invalid due to comma after number in deck",
			r: &room{
				name: "Hola",
				deck: "1,2,",
			},
			wantErrors: map[string]string{
				"deck": fmt.Sprintf("must match %s", deckRegex.String()),
			},
		},
		{
			name: "invalid due to too long entry in deck",
			r: &room{
				name: "Hola",
				deck: "1,2222",
			},
			wantErrors: map[string]string{
				"deck": fmt.Sprintf("must match %s", deckRegex.String()),
			},
		},
		{
			name: "invalid due to name too long",
			r: &room{
				name: strings.Repeat("a", 21),
				deck: "1,2,3",
			},
			wantErrors: map[string]string{
				"name": "must not be more than 20 bytes long",
			},
		},
		{
			name: "invalid due to empty name",
			r: &room{
				name: "",
				deck: "1,2",
			},
			wantErrors: map[string]string{
				"name": "must be provided",
			},
		},
		{
			name: "invalid due to not unique deck entries",
			r: &room{
				name: "Hello World",
				deck: "1,2,2",
			},
			wantErrors: map[string]string{
				"deck": "must be unique",
			},
		},
		{
			name: "invalid due to invalid amount of elements (> 10)",
			r: &room{
				name: "Hello World",
				deck: "1,2,3,4,5,6,7,8,9,10,11",
			},
			wantErrors: map[string]string{
				"deck": "must be 10 elements maximum",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()

			validateRoom(v, tt.r)

			assert.DeepEqual(t, v.Errors, tt.wantErrors)
		})
	}
}
