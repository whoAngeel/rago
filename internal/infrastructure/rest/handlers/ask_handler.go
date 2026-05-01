package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type AskHandler struct {
	usecase *application.AskUsecase
	logger  ports.Logger
}

func NewAskHandler(uc *application.AskUsecase, log ports.Logger) *AskHandler {
	return &AskHandler{
		usecase: uc,
		logger:  log,
	}
}

type askRequest struct {
	Question string `json:"question" binding:"required,gte=1"`
}

type askResponse struct {
	Answer string `json:"answer"`
}

func (h *AskHandler) Ask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req askRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}

	h.logger.Debug("BODY", "body", req)

	userID := c.GetInt("user_id")
	answer, err := h.usecase.Execute(ctx, userID, req.Question)
	if err != nil {
		h.logger.Error("error asking", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error asking",
		})
		return
	}

	var res askResponse
	res.Answer = answer

	c.JSON(http.StatusOK, askResponse{
		Answer: res.Answer,
	})
	return
}
