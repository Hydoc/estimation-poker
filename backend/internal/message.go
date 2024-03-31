package internal

const (
	guess    = "guess"
	newRound = "new-round"
	estimate = "estimate"
	lockRoom = "lock-room"
)

type messageDTO map[string]interface{}

type message interface {
	ToJson() messageDTO
}

type clientMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type join struct{}

type leave struct{}

type resetRound struct{}

type developerGuessed struct{}

type roomLocked struct{}

type everyoneGuessed struct{}

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

func (rLocked roomLocked) ToJson() messageDTO {
	return map[string]any{
		"type": "room-locked",
	}
}

func (everyOneGuessed everyoneGuessed) ToJson() messageDTO {
	return map[string]interface{}{
		"type": "everyone-guessed",
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

func newDeveloperGuessed() developerGuessed {
	return developerGuessed{}
}

func newEveryoneGuessed() everyoneGuessed {
	return everyoneGuessed{}
}

func newResetRound() resetRound {
	return resetRound{}
}

func newYouGuessed(guess int) youGuessed {
	return youGuessed{
		guess: guess,
	}
}
