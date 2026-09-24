package main

import (
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type issue struct {
	Id    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

func newIssue(title string) *issue {
	return &issue{
		Id:    uuid.New(),
		Title: title,
	}
}

type room struct {
	issuesMu sync.Mutex
	issues   []*issue

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
	r.issuesMu.Lock()
	defer r.issuesMu.Unlock()
	r.issues = append(r.issues, issue)
}

func (r *room) addSubscriber(s *subscriber) {
	r.subscribersMu.Lock()
	defer r.subscribersMu.Unlock()
	r.subscribers[s] = struct{}{}
}

func (r *room) deleteSubscriber(s *subscriber) {
	r.subscribersMu.Lock()
	defer r.subscribersMu.Unlock()
	delete(r.subscribers, s)
}

func (r *room) Issues() []*issue {
	r.issuesMu.Lock()
	defer r.issuesMu.Unlock()
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

func newRoom() *room {
	return &room{
		issues:      make([]*issue, 0),
		subscribers: make(map[*subscriber]struct{}),
	}
}
