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
	ssePkg "github.com/whoAngeel/rago/internal/infrastructure/sse"
	"github.com/whoAngeel/rago/internal/infrastructure/storage"
	"github.com/whoAngeel/rago/internal/worker"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const DefaultSystemPrompt = `Eres un asistente experto que responde preguntas basándose ÚNICAMENTE en la sección CONTEXTO proporcionada.
Instrucciones:
1. Usa solo la información en la sección CONTEXTO para responder.
2. Si el CONTEXTO no tiene información suficiente, responde: "No tengo información suficiente en tus documentos para responder a esto."
3. No inventes ni uses conocimiento general.
4. Si mencionas datos, cita las fuentes proporcionadas.`

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	log := logger.New(cfg.Env)

	log.Info("App starting...", "mode", cfg.Env, "port", cfg.Port)

	gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		DSN:                  cfg.DatabaseUrl,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
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
		&domain.SystemConfig{},
		&domain.ChatMessage{},
		&domain.ChatSession{},
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

	// seed system prompt
	gormDB.FirstOrCreate(&domain.SystemConfig{
		Key:   "system_prompt",
		Value: DefaultSystemPrompt,
	}, &domain.SystemConfig{Key: "system_prompt"})

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

	userRepo := postgres.NewUserRepository(gormDB)
	sessionRepo := postgres.NewSessionRepository(gormDB)
	docRepo := postgres.NewDocumentRepository(gormDB)
	chatRepo := postgres.NewChatRepository(gormDB)
	systemRepo := postgres.NewSystemConfigRepository(gormDB)

	ingestUC := application.NewIngestUsecase(vStore, embedder, log, *cfg)

	sseManager := ssePkg.NewManager()

	chatUC := application.NewChatUsecase(
		chatRepo,
		systemRepo,
		vStore,
		embedder,
		llm,
		sseManager,
		log,
		cfg.ChatHistoryLimit,
		cfg.QdrantCollection,
		cfg.ContextWindowLimit)

	parserRegistry := parserpkg.NewRegistry()
	parserRegistry.Register("text/plain", parserpkg.NewPlainTextAdapter())
	parserRegistry.Register("text/csv", parserpkg.NewCSVParser())
	parserRegistry.Register("application/json", parserpkg.NewJSONParser())
	parserRegistry.Register("application/vnd.openxmlformats-officedocument.wordprocessingml.document", parserpkg.NewDOCXParser())
	parserRegistry.Register("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", parserpkg.NewXLSXParser())
	parserRegistry.Register("application/pdf", parserpkg.NewPDFParser())

	chunker := cnunkerPkg.NewFixedChunker(1000, 200)

	worker := worker.NewIngestWorker(docRepo, minio, parserRegistry, chunker, embedder, ingestUC, sseManager, 10*time.Second, 3, 3, *cfg)

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
		ChatHandler: handlers.NewChatHandler(
			chatUC,
			log,
			*cfg,
		),
		SSEHandler: handlers.NewSSEHandler(sseManager, log),
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
