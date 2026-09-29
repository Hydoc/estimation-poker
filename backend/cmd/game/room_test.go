package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Hydoc/estimation-poker/backend/internal/assert"
	"github.com/Hydoc/estimation-poker/backend/internal/validator"
)

func Test_room_publish(t *testing.T) {
	t.Run("publish a message correctly", func(t *testing.T) {
		sub := &subscriber{
			messages:  make(chan outgoingMessage, 1),
			closeSlow: func() {},
		}
		r := &room{
			subscribers: map[*subscriber]struct{}{
				sub: {},
			},
		}
		r.addSubscriber(sub)

		msgToPublish := outgoingMessage{
			Type: "Test",
		}
		r.publish(msgToPublish)
		got := <-sub.messages
		assert.DeepEqual(t, got, msgToPublish)
	})

	t.Run("should call closeSlow for slow subscriber", func(t *testing.T) {
		closeCalled := make(chan struct{})
		sub := &subscriber{
			messages: make(chan outgoingMessage),
			closeSlow: func() {
				close(closeCalled)
			},
		}
		r := &room{
			subscribers: map[*subscriber]struct{}{
				sub: {},
			},
		}
		r.addSubscriber(sub)

		r.publish(outgoingMessage{
			Type: "Test",
		})

		select {
		case <-closeCalled:
		// Test is ok
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timed out waiting for closeSlow")
		}
	})
}

func Test_room_addIssue(t *testing.T) {
	r := newRoom("Test", "1,2,3,4,5")
	i := newIssue("Test Issue")
	r.addIssue(i)
	assert.DeepEqual(t, r.Issues(), []*issue{i})
}

func Test_room_deleteSubscriber(t *testing.T) {
	tests := []struct {
		name             string
		setup            func() (*room, *subscriber)
		wantRoomDeletion bool
	}{
		{
			name: "should delete room when all subscribers are gone",
			setup: func() (*room, *subscriber) {
				s := &subscriber{}

				r := &room{
					subscribers: map[*subscriber]struct{}{
						s: {},
					},
				}

				return r, s
			},
			wantRoomDeletion: true,
		},
		{
			name: "should keep room when one subscriber leaves",
			setup: func() (*room, *subscriber) {
				firstSubscriber := &subscriber{}
				secondSubscriber := &subscriber{}

				r := &room{
					subscribers: map[*subscriber]struct{}{
						firstSubscriber:  {},
						secondSubscriber: {},
					},
				}

				return r, firstSubscriber
			},
			wantRoomDeletion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, s := tt.setup()
			got := r.deleteSubscriber(s)
			assert.Equal(t, got, tt.wantRoomDeletion)
		})
	}
}

func Test_room_State(t *testing.T) {
	sub := newSubscriber("Tester", nil)
	r := newRoom("Test", "1,2,3,4,5")
	r.addSubscriber(sub)
	want := map[string]any{
		"players": []*subscriber{sub},
		"room": envelope{
			"name":   "Test",
			"deck":   "1,2,3,4,5",
			"issues": make([]*issue, 0),
		},
	}
	assert.DeepEqual(t, r.State(), want)
}

func Test_validateRoom(t *testing.T) {
	tests := []struct {
		name       string
		r          *room
		wantErrors map[string]string
	}{
		{
			name: "should be a valid room",
			r: &room{
				name: strings.Repeat("a", 20),
				deck: "1,a2,333,E4,5",
			},
			wantErrors: map[string]string{},
		},
		{
			name: "should be invalid due to single number as deck",
			r: &room{
				name: "Hola",
				deck: "1",
			},
			wantErrors: map[string]string{
				"deck": fmt.Sprintf("must match %s", deckRegex.String()),
			},
		},
		{
			name: "should be invalid due to comma after number in deck",
			r: &room{
				name: "Hola",
				deck: "1,2,",
			},
			wantErrors: map[string]string{
				"deck": fmt.Sprintf("must match %s", deckRegex.String()),
			},
		},
		{
			name: "should be invalid due to too long entry in deck",
			r: &room{
				name: "Hola",
				deck: "1,2222",
			},
			wantErrors: map[string]string{
				"deck": fmt.Sprintf("must match %s", deckRegex.String()),
			},
		},
		{
			name: "should be invalid due to name too long",
			r: &room{
				name: strings.Repeat("a", 21),
				deck: "1,2,3",
			},
			wantErrors: map[string]string{
				"name": "must not be more than 20 bytes long",
			},
		},
		{
			name: "should be invalid due to empty name",
			r: &room{
				name: "",
				deck: "1,2",
			},
			wantErrors: map[string]string{
				"name": "must be provided",
			},
		},
		{
			name: "should be invalid due to not unique deck entries",
			r: &room{
				name: "Hello World",
				deck: "1,2,2",
			},
			wantErrors: map[string]string{
				"deck": "must be unique",
			},
		},
		{
			name: "should be invalid due to invalid amount of elements (> 10)",
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
