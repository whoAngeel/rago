package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	cnunkerPkg "github.com/whoAngeel/rago/internal/infrastructure/chunker"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/logger"
	"github.com/whoAngeel/rago/internal/infrastructure/openrouter"
	parserpkg "github.com/whoAngeel/rago/internal/infrastructure/parser"
	"github.com/whoAngeel/rago/internal/infrastructure/postgres"
	"github.com/whoAngeel/rago/internal/infrastructure/qdrant"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
	"github.com/whoAngeel/rago/internal/infrastructure/rest/handlers"
	"github.com/whoAngeel/rago/internal/infrastructure/rest/middleware"
	"github.com/whoAngeel/rago/internal/infrastructure/storage"
	"github.com/whoAngeel/rago/internal/worker"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	log := logger.New(cfg.Env)

	log.Info("App starting...", "mode", cfg.Env, "port", cfg.Port)

	gormDB, err := gorm.Open(gormPostgres.Open(cfg.DatabaseUrl), &gorm.Config{
		// Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		log.Fatal("error connecting to database", "error", err)
	}
	models := []interface{}{
		&domain.Role{},
		&domain.User{},
		&domain.Session{},
		&domain.Document{},
		&domain.ProcessingStep{},
	}
	for _, model := range models {
		if err := gormDB.AutoMigrate(model); err != nil {
			log.Warn("migration warning", "model", fmt.Sprintf("%T", model), "error", err)
		}
	}

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

	minio, err := storage.NewMinioAdapter(
		cfg.MinioEndpoint,
		cfg.MinioRootUser,
		cfg.MinioRootPass,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatal("error initializing minio storage", "error", err)
	}
	log.Info("MINIO started")

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
	ingestUC := application.NewIngestUsecase(vStore, embedder, log, *cfg)

	userRepo := postgres.NewUserRepository(gormDB)
	sessionRepo := postgres.NewSessionRepository(gormDB)
	docRepo := postgres.NewDocumentRepository(gormDB)

	parser := parserpkg.NewPlainTextAdapter()
	chunker := cnunkerPkg.NewFixedChunker(1000, 200)

	worker := worker.NewIngestWorker(docRepo, minio, parser, chunker, embedder, ingestUC, 10*time.Second, 3, 3, *cfg)

	router := handlers.NewRouter(log, &handlers.Handlers{
		AskHandler: handlers.NewAskHandler(
			application.NewAskUsecase(vStore, llm, log, embedder, cfg),
			log,
		),
		AuthHandler: handlers.NewAuthHandler(
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
		DocumentHandler: handlers.NewDocumentHandler(
			application.NewIngestDocumentUsecase(
				docRepo,
				minio,
				ingestUC,
			),
			log,
			*cfg,
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

	go worker.Start(ctx)
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
