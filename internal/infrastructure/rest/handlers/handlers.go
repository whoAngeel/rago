package handlers

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/rest/middleware"
)

type Handlers struct {
	AskHandler *AskHandler
	// IngestHandler   *IngestHandler
	AuthHandler     *AuthHandler
	DocumentHandler *DocumentHandler
	ChatHandler     *ChatHandler
}

func NewRouter(logger ports.Logger, handlers *Handlers) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger))
	r.Use(cors.Default())
	r.Use(requestid.New(
		requestid.WithCustomHeaderStrKey("X-Request-ID"),
	))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	setupRoutes(
		r,
		handlers.AskHandler,
		// handlers.IngestHandler,
		handlers.AuthHandler,
		handlers.DocumentHandler,
		handlers.ChatHandler,
	)

	return r
}

func setupRoutes(
	router *gin.Engine,
	askHandler *AskHandler,
	// ingestHandler *IngestHandler,
	authHandler *AuthHandler,
	docHandler *DocumentHandler,
	chatHandler *ChatHandler,
) {
	v1 := router.Group("/api/v1")
	{

		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)
		}
		protected := v1.Group("")
		{
			protected.Use(middleware.AuthMiddleware())
			protected.POST("/ask", askHandler.Ask)
			// protected.POST("/ingest", ingestHandler.Ingest)
			documentGroup := protected.Group("/documents")
			{
				documentGroup.GET("/", docHandler.List)
				documentGroup.POST("/", docHandler.Upload)
				documentGroup.DELETE("/:id", docHandler.Delete)
			}

			chatGroup := protected.Group("/chats")
			{
				chatGroup.POST("/send", chatHandler.SendMessage)
				// chatGroup.GET("/", chatHandler.List)
				// chatGroup.GET("/:id", chatHandler.GetByID)
				// chatGroup.POST("/", chatHandler.Create)
				// chatGroup.DELETE("/:id", chatHandler.Delete)
			}
		}

	}
}
