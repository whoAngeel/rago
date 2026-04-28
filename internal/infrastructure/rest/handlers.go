package rest

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type Handlers struct {
	AskHandler    *AskHandler
	IngestHandler *IngestHandler
	AuthHandler   *AuthHandler
}

func NewRouter(logger ports.Logger, handlers *Handlers) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(logger))
	r.Use(cors.Default())
	r.Use(requestid.New(
		requestid.WithCustomHeaderStrKey("X-Request-ID"),
	))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	setupRoutes(r, handlers.AskHandler, handlers.IngestHandler, handlers.AuthHandler)

	return r
}

func setupRoutes(
	router *gin.Engine,
	askHandler *AskHandler,
	ingestHandler *IngestHandler,
	authHandler *AuthHandler,
) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/ask", askHandler.Ask)
		v1.POST("/ingest", ingestHandler.Ingest)

		authGroup := router.Group("auth/")
		{
			authGroup.POST("/register", authHandler.Register)
		}
	}
}
