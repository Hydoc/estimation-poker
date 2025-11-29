package internal

const (
	join             = "join"
	leave            = "leave"
	Guess            = "guess"
	NewRound         = "new-round"
	Estimate         = "estimate"
	LockRoom         = "lock-room"
	OpenRoom         = "open-room"
	SkipRound        = "skip"
	Reveal           = "reveal"
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

type SkipRoundPayload struct {
	client *Client
}

type NewRoundPayload struct {
	client *Client
}

type LockRoomPayload struct {
	client   *Client
	key      string
	password string
}

type OpenRoomPayload struct {
	client *Client
	key    string
}

type EstimatePayload struct {
	client *Client
	ticket string
}

type GuessPayload struct {
	client *Client
	guess  int
}

type RevealPayload struct {
	client *Client
}

func NewJoin() *Message {
	return &Message{
		Type: join,
	}
}

func newEstimate(ticket string) *Message {
	return &Message{
		Type: Estimate,
		Data: ticket,
	}
}

func newLeave(name string) *Message {
	return &Message{
		Type: leave,
		Data: name,
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
		Type: Reveal,
		Data: out,
	}
}

func newNewRound() *Message {
	return &Message{
		Type: NewRound,
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
