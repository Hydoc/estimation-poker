package internal

const (
	Guess    = "guess"
	NewRound = "new-round"
)

type messageDTO map[string]interface{}

type message interface {
	ToJson() messageDTO
}

type incomingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type join struct{}

type leave struct{}

type resetRound struct{}

type developerGuessed struct{}

type everyoneGuessed struct{}

type youGuessed struct {
	guess int
}

func (leave leave) ToJson() messageDTO {
	return map[string]interface{}{
		"type": "leave",
	}
}
func (incMessage incomingMessage) ToJson() messageDTO {
	return map[string]interface{}{
		"type": incMessage.Type,
		"data": incMessage.Data,
	}
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
