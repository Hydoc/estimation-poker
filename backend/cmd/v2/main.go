package main

import (
	"log/slog"
	"os"

	"github.com/Hydoc/guess-dev/backend/internal/game"
	"github.com/Hydoc/guess-dev/backend/internal/message"
)

type application struct {
	logger *slog.Logger
	rooms  game.Rooms
	bus    message.MessageBus
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &application{
		logger: logger,
		rooms:  game.NewRoomModel(),
		bus:    message.NewBus(),
	}

	err := app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
