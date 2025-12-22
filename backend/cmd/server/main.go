package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Hydoc/go-message"

	"github.com/Hydoc/guess-dev/backend/internal"
)

const (
	defaultGuesses            = "1,2,3,4,5"
	defaultGuessesDescription = "Up to 4h,Up to 8h,Up to 3 days,Up to 5 days,More than 5 days"
	version                   = "1.20.0"
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
	rooms       map[internal.RoomId]*internal.Room
	destroyRoom chan internal.RoomId
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
		rooms:       make(map[internal.RoomId]*internal.Room),
		destroyRoom: make(chan internal.RoomId),
		bus:         createBus(),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go app.listenForRoomDestroy(ctx)

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

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
