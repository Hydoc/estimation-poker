package internal

const (
	join             = "join"
	leave            = "leave"
	guess            = "guess"
	newRound         = "new-round"
	estimate         = "estimate"
	lockRoom         = "lock-room"
	openRoom         = "open-room"
	skipRound        = "skip"
	reveal           = "reveal"
	roomLocked       = "room-locked"
	roomOpened       = "room-opened"
	developerGuessed = "developer-guessed"
	everyoneDone     = "everyone-done"
	developerSkipped = "developer-skipped"
	youSkipped       = "you-skipped"
	youGuessed       = "you-guessed"
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func NewJoin() *Message {
	return &Message{
		Type: join,
	}
}

func newLeave() *Message {
	return &Message{
		Type: leave,
	}
}

func newRoomLocked() *Message {
	return &Message{
		Type: roomLocked,
	}
}

func newRoomOpened() *Message {
	return &Message{
		Type: roomOpened,
	}
}

func newDeveloperGuessed() *Message {
	return &Message{
		Type: developerGuessed,
	}
}

func newEveryoneIsDone() *Message {
	return &Message{
		Type: everyoneDone,
	}
}

func newReveal(clients map[*Client]bool) *Message {
	out := []map[string]any{}
	for client := range clients {
		if client.Role == Developer {
			out = append(out, client.asReveal())
		}
	}

	return &Message{
		Type: reveal,
		Data: out,
	}
}

func newNewRound() *Message {
	return &Message{
		Type: newRound,
	}
}

func newDeveloperSkipped() *Message {
	return &Message{
		Type: developerSkipped,
	}
}

func newYouSkipped() *Message {
	return &Message{
		Type: youSkipped,
	}
}

func newYouGuessed(guess int) *Message {
	return &Message{
		Type: youGuessed,
		Data: guess,
	}
}
