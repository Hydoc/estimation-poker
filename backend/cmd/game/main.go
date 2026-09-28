package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	err := serve(logger, &cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func serve(logger *slog.Logger, config *config) error {
	server := newGameServer(logger, config, initMessageHandlerRegistry())

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      server.routes(),
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	shutdownErr := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		logger.Info("stopping server", "addr", srv.Addr, "signal", s.String())
		shutdownErr <- srv.Shutdown(context.Background())
	}()

	logger.Info("starting server", "addr", srv.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	logger.Info("stopped server")
	return nil
}

func initMessageHandlerRegistry() *messageHandlerRegistry {
	handlerRegistry := newMessageHandlerRegistry()
	handlerRegistry.register(issueAdd, handleIssueAddMessage)
	handlerRegistry.register(roomJoin, handleRoundJoinMessage)
	handlerRegistry.register(roomLeave, handleRoundLeaveMessage)
	return handlerRegistry
}
