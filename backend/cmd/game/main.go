package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	port int
	env  string
}

func main() {
	var cfg config
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	flag.IntVar(&cfg.port, "port", 8080, "port to listen on")
	flag.Parse()

	err := serve(logger, &cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func serve(logger *slog.Logger, config *config) error {
	server := newGameServer(logger, initMessageHandlerRegistry())

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      server.routes(),
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		logger.Info("caught signal", "signal", s.String())

		os.Exit(0)
	}()

	logger.Info("starting server", "addr", srv.Addr)

	return srv.ListenAndServe()
}

func initMessageHandlerRegistry() *messageHandlerRegistry {
	handlerRegistry := newMessageHandlerRegistry()
	handlerRegistry.register(issueAdd, handleIssueAddMessage)
	handlerRegistry.register(roundJoin, handleRoundJoinMessage)
	handlerRegistry.register(roundLeave, handleRoundLeaveMessage)
	return handlerRegistry
}
