package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type IngestHandler struct {
	usecase *application.IngestUsecase
	logger  ports.Logger
}

func NewIngestHandler(uc *application.IngestUsecase, log ports.Logger) *IngestHandler {
	return &IngestHandler{
		usecase: uc,
		logger:  log,
	}
}

type IngestRequest struct {
	Filename string `json:"filename" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

func (h *IngestHandler) Ingest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename and content required"})
		return
	}

	if err := h.usecase.Execute(ctx, req.Filename, req.Content); err != nil {
		h.logger.Error("error ingesting", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error ingesting document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document ingested"})
}
