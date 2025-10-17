package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Hydoc/go-message"
	"github.com/Hydoc/guess-dev/backend/internal"
)

const (
	defaultGuesses            = "1,2,3,4,5"
	defaultGuessesDescription = "Bis zu 4 Std.,Bis zu 8 Std.,Bis zu 3 Tagen,Bis zu 5 Tagen,Mehr als 5 Tage"
)

func createBus() message.Bus {
	bus := message.NewBus()
	bus.Register(internal.SkipRound, internal.HandleSkipRound)
	bus.Register(internal.Estimate, internal.HandleEstimate)
	bus.Register(internal.Guess, internal.HandleGuess)
	bus.Register(internal.NewRound, internal.HandleNewRound)
	bus.Register(internal.Reveal, internal.HandleReveal)
	bus.Register(internal.LockRoom, internal.HandleLockRoom)
	bus.Register(internal.OpenRoom, internal.HandleOpenRoom)
	return bus
}

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

	app := &application{
		logger:      logger,
		guessConfig: config,
		rooms:       make(map[internal.RoomId]*internal.Room),
		destroyRoom: make(chan internal.RoomId),
		bus:         createBus(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go app.listenForRoomDestroy(ctx)
	router := app.Routes()
	logger.Error(http.ListenAndServe(":8080", router).Error())
}
