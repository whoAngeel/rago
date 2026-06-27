package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/rest/middleware"
)

type ServicePingFn func(ctx context.Context) error

type HealthChecks struct {
	Postgres ServicePingFn
	Qdrant   ServicePingFn
	Storage  ServicePingFn
}

type serviceResult struct {
	Status    string  `json:"status"`
	LatencyMs float64 `json:"latency_ms"`
	Error     string  `json:"error,omitempty"`
}

type Handlers struct {
	AskHandler           *AskHandler
	AuthHandler          *AuthHandler
	ChatHandler          *ChatHandler
	ConfigHandler        *ConfigHandler
	DocumentHandler      *DocumentHandler
	DocumentGroupHandler *DocumentGroupHandler
	PublicGroupHandler   *PublicGroupHandler
	SSEHandler           *SSEHandler
	UserHandler          *UserHandler
}

func NewRouter(logger ports.Logger, handlers *Handlers, checks HealthChecks) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger))
	r.Use(cors.Default())
	r.Use(requestid.New(
		requestid.WithCustomHeaderStrKey("X-Request-ID"),
	))

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		type named struct {
			name string
			fn   ServicePingFn
		}
		services := []named{
			{"postgres", checks.Postgres},
			{"qdrant", checks.Qdrant},
			{"storage", checks.Storage},
		}

		results := make(map[string]serviceResult, len(services))
		var mu sync.Mutex
		var wg sync.WaitGroup
		allOK := true

		for _, svc := range services {
			svc := svc
			wg.Add(1)
			go func() {
				defer wg.Done()
				start := time.Now()
				err := svc.fn(ctx)
				lat := float64(time.Since(start).Milliseconds())
				r := serviceResult{Status: "ok", LatencyMs: lat}
				if err != nil {
					r.Status = "down"
					r.Error = err.Error()
					mu.Lock()
					allOK = false
					mu.Unlock()
				}
				mu.Lock()
				results[svc.name] = r
				mu.Unlock()
			}()
		}
		wg.Wait()

		status := "ok"
		code := http.StatusOK
		if !allOK {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		c.JSON(code, gin.H{"status": status, "services": results})
	})

	setupRoutes(r, handlers)

	return r
}

func setupRoutes(router *gin.Engine, h *Handlers) {
	v1 := router.Group("/api/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.AuthHandler.Register)
		auth.POST("/login", h.AuthHandler.Login)
		auth.POST("/refresh", h.AuthHandler.Refresh)
		auth.POST("/logout", h.AuthHandler.Logout)
	}

	public := v1.Group("/public")
	{
		public.GET("/groups/:slug", h.PublicGroupHandler.GetGroupInfo)
		public.POST("/groups/:slug/chat", h.PublicGroupHandler.Chat)
		public.GET("/groups/:slug/documents/:doc_id/download", h.PublicGroupHandler.DownloadDocument)
		public.GET("/groups/:slug/documents/:doc_id/view", h.PublicGroupHandler.DownloadDocument)
	}

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/users/me", h.UserHandler.Me)
		protected.GET("/users/me/stats", h.UserHandler.Stats)

		protected.POST("/ask", h.AskHandler.Ask)
		protected.GET("/stream", h.SSEHandler.Stream)

		config := protected.Group("/config")
		{
			config.GET("/system-prompt", h.ConfigHandler.GetSystemPrompt)
			config.PUT("/system-prompt", h.ConfigHandler.UpdateSystemPrompt)
		}

		documents := protected.Group("/documents")
		{
			documents.GET("/select", h.DocumentHandler.ListSelect)
			documents.GET("/", h.DocumentHandler.List)
			documents.POST("/", h.DocumentHandler.Upload)
			documents.DELETE("/:id", h.DocumentHandler.Delete)
			documents.GET("/:id/steps", h.DocumentHandler.Steps)
			documents.GET("/:id/health", h.DocumentHandler.Health)
			documents.POST("/:id/reprocess", h.DocumentHandler.Reprocess)
		}

		groups := protected.Group("/groups")
		{
			groups.POST("/", h.DocumentGroupHandler.Create)
			groups.GET("/", h.DocumentGroupHandler.List)
			groups.GET("/:id", h.DocumentGroupHandler.Get)
			groups.GET("/:id/usage", h.DocumentGroupHandler.GetUsage)
			groups.PATCH("/:id", h.DocumentGroupHandler.Update)
			groups.DELETE("/:id", h.DocumentGroupHandler.Delete)
			groups.POST("/:id/documents", h.DocumentGroupHandler.AddDocuments)
			groups.DELETE("/:id/documents/:doc_id", h.DocumentGroupHandler.RemoveDocument)
		}

		chats := protected.Group("/chats")
		{
			chats.POST("/send", h.ChatHandler.SendMessage)
			chats.POST("/send-stream", h.ChatHandler.SendStream)
			chats.GET("/", h.ChatHandler.ListSessions)
			chats.GET("/:id", h.ChatHandler.GetSession)
			chats.PATCH("/:id", h.ChatHandler.UpdateSessionTittle)
			chats.DELETE("/:id", h.ChatHandler.Delete)
		}
	}
}
