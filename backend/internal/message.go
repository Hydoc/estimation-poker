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

type message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func newJoin() *message {
	return &message{
		Type: join,
	}
}

func newLeave() *message {
	return &message{
		Type: leave,
	}
}

func newRoomLocked() *message {
	return &message{
		Type: roomLocked,
	}
}

func newRoomOpened() *message {
	return &message{
		Type: roomOpened,
	}
}

func newDeveloperGuessed() *message {
	return &message{
		Type: developerGuessed,
	}
}

func newEveryoneIsDone() *message {
	return &message{
		Type: everyoneDone,
	}
}

func newReveal(clients map[*Client]bool) *message {
	out := []map[string]any{}
	for client := range clients {
		if client.Role == Developer {
			out = append(out, client.asReveal())
		}
	}

	return &message{
		Type: reveal,
		Data: out,
	}
}

func newNewRound() *message {
	return &message{
		Type: newRound,
	}
}

func newDeveloperSkipped() *message {
	return &message{
		Type: developerSkipped,
	}
}

func newYouSkipped() *message {
	return &message{
		Type: youSkipped,
	}
}

func newYouGuessed(guess int) *message {
	return &message{
		Type: youGuessed,
		Data: guess,
	}
}
