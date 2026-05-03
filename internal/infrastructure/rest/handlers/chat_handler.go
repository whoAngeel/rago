package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
)

type ChatHandler struct {
	useCase *application.ChatUsecase
	logger  ports.Logger
	config  config.Config
}

func NewChatHandler(uc *application.ChatUsecase, logger ports.Logger, config config.Config) *ChatHandler {
	return &ChatHandler{
		useCase: uc,
		logger:  logger,
		config:  config,
	}
}

type SendMessageRequest struct {
	SessionID *int   `json:"session_id,omitempty"`
	Question  string `json:"question" validate:"required"`
}

type SendMessageResponse struct {
	Answer    string               `json:"answer"`
	Sources   []application.Source `json:"sources"`
	SessionID int                  `json:"session_id"`
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body")
		rest.RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userId := c.GetInt("user_id")
	answer, sources, sessionId, err := h.useCase.SendMessage(ctx, userId, req.SessionID, req.Question)
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Failed to send message", err.Error())
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Answer:    answer,
		Sources:   sources,
		SessionID: sessionId,
	})
}
