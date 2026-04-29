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
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/logger"
	"github.com/whoAngeel/rago/internal/infrastructure/openrouter"
	"github.com/whoAngeel/rago/internal/infrastructure/postgres"
	"github.com/whoAngeel/rago/internal/infrastructure/qdrant"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
	"github.com/whoAngeel/rago/internal/infrastructure/rest/middleware"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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

	gormDB, err := gorm.Open(gormPostgres.Open(cfg.DatabaseUrl), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	gormDB.AutoMigrate(&domain.User{}, &domain.Session{})

	// seed roles
	var roleCount int64
	gormDB.Model(&domain.Role{}).Count(&roleCount)
	if roleCount == 0 {
		gormDB.Create(&[]domain.Role{
			{Name: "admin"}, {Name: "viewer"}, {Name: "editor"},
		})
	}

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

	// inicializar repositories
	userRepo := postgres.NewUserRepository(gormDB)
	sessionRepo := postgres.NewSessionRepository(gormDB)

	router := rest.NewRouter(log, &rest.Handlers{
		AskHandler: rest.NewAskHandler(
			application.NewAskUsecase(vStore, llm, log, embedder, cfg),
			log,
		),
		IngestHandler: rest.NewIngestHandler(
			application.NewIngestUsecase(vStore, embedder, log, *cfg),
			log,
		),
		AuthHandler: rest.NewAuthHandler(
			application.NewAuthUseCase(
				userRepo,
				sessionRepo,
				cfg.Secret,
				log,
				cfg.AccessTokenExpiration,
				cfg.RefreshTokenExpiration,
			),
			log,
		),
	})
	server := rest.NewServer(cfg.Host, cfg.Port, router, log)
	middleware.InitJWTSecret(cfg.Secret)

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
