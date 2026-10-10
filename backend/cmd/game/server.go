package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (srv *gameServer) serve() error {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.config.port),
		Handler:      srv.routes(),
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		ErrorLog:     slog.NewLogLogger(srv.logger.Handler(), slog.LevelError),
	}

	shutdownErr := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		srv.logger.Info("stopping server", "addr", httpServer.Addr, "signal", s.String())
		shutdownErr <- httpServer.Shutdown(context.Background())
	}()

	srv.logger.Info("starting server", "addr", httpServer.Addr)

	err := httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	srv.logger.Info("stopped server")
	return nil
}
