package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Hydoc/estimation-poker/backend/internal/validator"
	"github.com/google/uuid"
)

var (
	maxAllowedCardsPerDeck = 10
	deckRegex              = regexp.MustCompile("^[^,]{1,3}(,[^,]{1,3})+$")
)

type issue struct {
	Id     uuid.UUID `json:"id"`
	Title  string    `json:"title"`
	Effort int       `json:"effort"`
}

func newIssue(title string) *issue {
	return &issue{
		Id:     uuid.New(),
		Title:  title,
		Effort: -1,
	}
}

type room struct {
	mu     sync.RWMutex
	issues []*issue

	name string
	deck string

	inProgress atomic.Bool

	subscribersMu sync.RWMutex
	subscribers   map[*subscriber]struct{}
}

func (r *room) publish(msg outgoingMessage) {
	r.subscribersMu.RLock()
	subs := make([]*subscriber, 0, len(r.subscribers))
	for s := range r.subscribers {
		subs = append(subs, s)
	}
	r.subscribersMu.RUnlock()

	for _, s := range subs {
		select {
		case s.messages <- msg:
		default:
			go s.closeSlow()
		}
	}
}

func (r *room) addIssue(issue *issue) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues = append(r.issues, issue)
}

func (r *room) addSubscriber(s *subscriber) {
	r.subscribersMu.Lock()
	defer r.subscribersMu.Unlock()
	r.subscribers[s] = struct{}{}
}

func (r *room) deleteSubscriber(s *subscriber) bool {
	r.subscribersMu.Lock()
	defer r.subscribersMu.Unlock()
	delete(r.subscribers, s)
	return len(r.subscribers) == 0
}

func (r *room) Issues() []*issue {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*issue, len(r.issues))
	copy(cp, r.issues)
	return cp
}

func (r *room) SetInProgress(value bool) {
	r.inProgress.Store(value)
}

func (r *room) InProgress() bool {
	return r.inProgress.Load()
}

func (r *room) SubscribersSlice() []*subscriber {
	r.subscribersMu.Lock()
	defer r.subscribersMu.Unlock()

	subscribers := make([]*subscriber, 0, len(r.subscribers))
	for s := range r.subscribers {
		subscribers = append(subscribers, s)
	}
	return subscribers
}

func (r *room) State() map[string]any {
	return map[string]any{
		"players": r.SubscribersSlice(),
		"room": envelope{
			"name":   r.name,
			"deck":   r.deck,
			"issues": r.Issues(),
		},
	}
}

func validateRoom(v *validator.Validator, r *room) {
	v.Check(r.name != "", "name", "must be provided")
	v.Check(len(r.name) <= maxAllowedCharsPerName, "name", "must not be more than 20 bytes long")

	v.Check(validator.Matches(r.deck, deckRegex), "deck", fmt.Sprintf("must match %s", deckRegex.String()))

	splitDeck := strings.Split(r.deck, ",")
	v.Check(validator.Unique(splitDeck), "deck", "must be unique")
	v.Check(len(splitDeck) <= maxAllowedCardsPerDeck, "deck", "must be 10 elements maximum")
}

func newRoom(name, deck string) *room {
	return &room{
		name:        name,
		issues:      make([]*issue, 0),
		deck:        deck,
		subscribers: make(map[*subscriber]struct{}),
	}
}
