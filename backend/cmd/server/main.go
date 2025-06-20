package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Hydoc/guess-dev/backend/internal"
)

const (
	defaultGuesses            = "1,2,3,4,5"
	defaultGuessesDescription = "Bis zu 4 Std.,Bis zu 8 Std.,Bis zu 3 Tagen,Bis zu 5 Tagen,Mehr als 5 Tage"
)

func main() {
	var possibleGuesses, possibleGuessesDescription string
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	possibleGuesses, ok := os.LookupEnv("POSSIBLE_GUESSES")
	possibleGuessesDescription, okDesc := os.LookupEnv("POSSIBLE_GUESSES_DESC")
	if !ok {
		logger.Info("can not find env POSSIBLE_GUESSES")
		possibleGuesses = defaultGuesses
	}
	if !okDesc {
		logger.Info("can not find env POSSIBLE_GUESSES_DESC")
		possibleGuessesDescription = defaultGuessesDescription
	}
	logger.Info(fmt.Sprintf("using possible guesses %s", possibleGuesses))
	logger.Info(fmt.Sprintf("using possible guesses description %s", possibleGuessesDescription))

	config, err := internal.NewGuessConfig(possibleGuesses, possibleGuessesDescription)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	app := internal.NewApplication(config, logger)
	go app.ListenForRoomDestroy()
	router := app.Routes()
	logger.Error(http.ListenAndServe(":8080", router).Error())
}
