package main

import (
	"flag"
	"log/slog"
	"os"
)

const version = "0.0.1"

type config struct {
	port int
	env  string
}

func main() {
	var cfg config
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	flag.IntVar(&cfg.port, "port", 8080, "port to listen on")
	flag.StringVar(&cfg.env, "env", "dev", "environment dev|test|prod")
	flag.Parse()

	srv := newGameServer(logger, &cfg, initMessageHandlerRegistry())
	err := srv.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func initMessageHandlerRegistry() *messageHandlerRegistry {
	handlerRegistry := newMessageHandlerRegistry()
	handlerRegistry.register(issueAdd, handleIssueAddMessage)
	handlerRegistry.register(roomJoin, handleRoundJoinMessage)
	handlerRegistry.register(roomLeave, handleRoundLeaveMessage)
	return handlerRegistry
}
