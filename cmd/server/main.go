package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/logger"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, _ := config.Load()
	log := logger.New(cfg.Env)

	log.Info("App starting", "mode", cfg.Env, "port", cfg.Port)

	router := rest.NewRouter(log)
	server := rest.NewServer(cfg.Host, cfg.Port, router, log)

	serverErr := make(chan error, 1)
	go func() {
		if err := server.Start(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()
	// SHUTDOWN SIGNAL HANDLING
	select {
	case err := <-serverErr:
		log.Fatal("server error", "err", err)
	case <-ctx.Done():
		log.Info("shutting signal received")
	}

	// GRACEFUL SHUTDOWN
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server forced to shutdown", "err", err)
	}

	log.Info("server exiting")
}
