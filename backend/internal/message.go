package internal

const (
	guess     = "guess"
	newRound  = "new-round"
	estimate  = "estimate"
	lockRoom  = "lock-room"
	openRoom  = "open-room"
	skipRound = "skip-round"
)

type messageDTO map[string]interface{}

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

type developerGuessed struct{}

type roomLocked struct{}

type roomOpened struct{}

type everyoneIsDone struct{}

type youGuessed struct {
	guess int
}

func (leave leave) ToJson() messageDTO {
	return map[string]interface{}{
		"type": "leave",
	}
}
func (incMessage clientMessage) ToJson() messageDTO {
	return map[string]interface{}{
		"type": incMessage.Type,
		"data": incMessage.Data,
	}
}

func (incMessage clientMessage) isEstimate() bool {
	return incMessage.Type == estimate
}

func (join join) ToJson() messageDTO {
	return map[string]interface{}{
		"type": "join",
	}
}

func (devGuessed developerGuessed) ToJson() messageDTO {
	return map[string]interface{}{
		"type": "developer-guessed",
	}
}

func (s skip) ToJson() messageDTO {
	return map[string]interface{}{
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
	return map[string]interface{}{
		"type": "everyone-done",
	}
}

func (resetRound resetRound) ToJson() messageDTO {
	return map[string]interface{}{
		"type": "reset-round",
	}
}

func (youGuessed youGuessed) ToJson() messageDTO {
	return map[string]interface{}{
		"type": "you-guessed",
		"data": youGuessed.guess,
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

func newSkip() skip {
	return skip{}
}

func newYouGuessed(guess int) youGuessed {
	return youGuessed{
		guess: guess,
	}
}
