package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/logger"
	"github.com/whoAngeel/rago/internal/infrastructure/openrouter"
	"github.com/whoAngeel/rago/internal/infrastructure/qdrant"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	log := logger.New(cfg.Env)

	log.Info("App starting", "mode", cfg.Env, "port", cfg.Port)

	// Inicializar servicios
	vStore, err := qdrant.NewQdrantAdapter(cfg.QdrantHost, cfg.QdrantPort)
	if err != nil {
		log.Fatal("error initializing qdrant", "error", err)
	}

	llm, err := openrouter.NewOpenRouterAdapter(
		cfg.OpenRouterKey, cfg.OpenRouterBaseUrl, cfg.Model,
	)
	if err != nil {
		log.Fatal("error initializing llm", "error", err)
	}

	embedder, err := openrouter.NewEmbedderAdapter(cfg.OpenRouterKey, cfg.OpenRouterBaseUrl, cfg.EmbeddingModel)
	if err != nil {
		log.Fatal("error initializing embedder", "error", err)
	}

	router := rest.NewRouter(log, &rest.Handlers{
		AskHandler: rest.NewAskHandler(
			application.NewAskUsecase(vStore, llm, log, embedder, cfg),
			log,
		),
		IngestHandler: rest.NewIngestHandler(
			application.NewIngestUsecase(vStore, embedder, log, *cfg),
			log,
		),
	})
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
