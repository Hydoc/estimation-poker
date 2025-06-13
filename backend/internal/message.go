package internal

const (
	guess     = "guess"
	newRound  = "new-round"
	estimate  = "estimate"
	lockRoom  = "lock-room"
	openRoom  = "open-room"
	skipRound = "skip"
	reveal    = "reveal"
)

type messageDTO map[string]any

type message interface {
	ToJson() messageDTO
}

type clientMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type skip struct{}

type join struct{}

type leave struct{}

type resetRound struct{}

type revealRound struct {
	clients map[*Client]bool
}

type developerGuessed struct{}

type roomLocked struct{}

type roomOpened struct{}

type everyoneIsDone struct{}

type youGuessed struct {
	guess int
}

type youSkipped struct{}

func (leave leave) ToJson() messageDTO {
	return map[string]any{
		"type": "leave",
	}
}
func (incMessage clientMessage) ToJson() messageDTO {
	return map[string]any{
		"type": incMessage.Type,
		"data": incMessage.Data,
	}
}

func (incMessage clientMessage) isEstimate() bool {
	return incMessage.Type == estimate
}

func (join join) ToJson() messageDTO {
	return map[string]any{
		"type": "join",
	}
}

func (devGuessed developerGuessed) ToJson() messageDTO {
	return map[string]any{
		"type": "developer-guessed",
	}
}

func (s skip) ToJson() messageDTO {
	return map[string]any{
		"type": "developer-skipped",
	}
}

func (rLocked roomLocked) ToJson() messageDTO {
	return map[string]any{
		"type": "room-locked",
	}
}

func (rOpened roomOpened) ToJson() messageDTO {
	return map[string]any{
		"type": "room-opened",
	}
}

func (everyOneGuessed everyoneIsDone) ToJson() messageDTO {
	return map[string]any{
		"type": "everyone-done",
	}
}

func (resetRound resetRound) ToJson() messageDTO {
	return map[string]any{
		"type": "reset-round",
	}
}

func (youGuessed youGuessed) ToJson() messageDTO {
	return map[string]any{
		"type": "you-guessed",
		"data": youGuessed.guess,
	}
}

func (youSkipped youSkipped) ToJson() messageDTO {
	return map[string]any{
		"type": "you-skipped",
	}
}

func (revealRound revealRound) ToJson() messageDTO {
	out := []map[string]any{}
	for client := range revealRound.clients {
		if client.Role == Developer {
			out = append(out, client.AsReveal())
		}
	}

	return map[string]any{
		"type": "reveal-round",
		"data": out,
	}
}

func newJoin() join {
	return join{}
}

func newLeave() leave {
	return leave{}
}

func newRoomLocked() roomLocked {
	return roomLocked{}
}

func newRoomOpened() roomOpened {
	return roomOpened{}
}

func newDeveloperGuessed() developerGuessed {
	return developerGuessed{}
}

func newEveryoneIsDone() everyoneIsDone {
	return everyoneIsDone{}
}

func newResetRound() resetRound {
	return resetRound{}
}

func newRevealRound(clients map[*Client]bool) revealRound {
	return revealRound{
		clients: clients,
	}
}

func newSkipRound() skip {
	return skip{}
}

func newYouSkipped() youSkipped {
	return youSkipped{}
}

func newYouGuessed(guess int) youGuessed {
	return youGuessed{
		guess: guess,
	}
}
