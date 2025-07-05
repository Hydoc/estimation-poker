package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type GuessConfig struct {
	Guesses []GuessConfigEntry
}

type GuessConfigEntry struct {
	Guess       int    `json:"guess"`
	Description string `json:"description"`
}

func NewGuessConfig(possibleGuesses, possibleGuessesDescription string) (*GuessConfig, error) {
	var guesses []GuessConfigEntry

	splitPossibleGuesses := strings.Split(possibleGuesses, ",")
	splitGuessesDesc := strings.Split(possibleGuessesDescription, ",")

	if len(splitPossibleGuesses) != len(splitGuessesDesc) {
		return nil, errors.New(fmt.Sprintf("error length for guesses and guesses desc is differrent (guesses = %d, guesses desc = %d)", len(splitPossibleGuesses), len(splitGuessesDesc)))
	}

	for i, guessStr := range splitPossibleGuesses {
		value, err := strconv.Atoi(guessStr)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error can not convert guess %s to int", guessStr))
		}

		guesses = append(guesses, GuessConfigEntry{
			Guess:       value,
			Description: splitGuessesDesc[i],
		})
	}

	return &GuessConfig{
		Guesses: guesses,
	}, nil
}
