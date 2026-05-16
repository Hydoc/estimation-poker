package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/Hydoc/go-message"
	"github.com/google/uuid"

	"github.com/Hydoc/estimation-poker/backend/internal"
)

const (
	defaultGuesses            = "1,2,3,4,5"
	defaultGuessesDescription = "Up to 4h,Up to 8h,Up to 3 days,Up to 5 days,More than 5 days"
	version                   = "1.21.0"
)

type config struct {
	port int
	env  string
}

type application struct {
	mu sync.Mutex

	config      config
	bus         message.Bus
	logger      *slog.Logger
	guessConfig *internal.GuessConfig
	rooms       map[uuid.UUID]*internal.Room
	destroyRoom chan uuid.UUID
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|test|production)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var possibleGuesses, possibleGuessesDescription string

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

	guessConfig, err := internal.NewGuessConfig(possibleGuesses, possibleGuessesDescription)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	app := &application{
		logger:      logger,
		config:      cfg,
		guessConfig: guessConfig,
		rooms:       make(map[uuid.UUID]*internal.Room),
		destroyRoom: make(chan uuid.UUID),
		bus:         internal.CreateBus(),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
