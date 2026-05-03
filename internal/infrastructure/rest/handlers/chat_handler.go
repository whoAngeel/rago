package handlers

import (
	"context"
	"net/http"
	"strconv"
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
	Question  string `json:"question" binding:"required,gte=1"`
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
		h.logger.Error("Failed to send message", "error", err)
		rest.RespondError(c, http.StatusInternalServerError, "Failed to send message", "")
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Answer:    answer,
		Sources:   sources,
		SessionID: sessionId,
	})
}

type ListSessionsResponse struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *ChatHandler) ListSessions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userId := c.GetInt("user_id")

	sessions, err := h.useCase.ListSessions(ctx, userId)
	if err != nil {
		h.logger.Error("Error getting sessions", "error", err)
		rest.RespondError(c, http.StatusInternalServerError, "Error getting sessions", "")
		return
	}

	var sessionsResponse []ListSessionsResponse
	for _, session := range sessions {
		sessionsResponse = append(sessionsResponse, ListSessionsResponse{
			ID:        int(session.ID),
			Title:     session.Title,
			UpdatedAt: session.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, sessionsResponse)
}

type MessageResponse struct {
	ID        uint      `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Sources   string    `json:"sources"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *ChatHandler) GetSession(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid session ID", err.Error())
		return
	}

	userID := c.GetInt("user_id")
	messages, err := h.useCase.GetSessionHistory(ctx, sessionID, userID)
	if err != nil {
		h.logger.Error("Error getting session messages", "error", err)
		rest.RespondError(c, http.StatusInternalServerError, "Error getting session messages", "")
		return
	}

	var response []MessageResponse
	for _, m := range messages {
		response = append(response, MessageResponse{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			Sources:   string(m.Sources),
			CreatedAt: m.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

type UpdateSessionTittleRequest struct {
	Title string `json:"title" binding:"required"`
}

func (h *ChatHandler) UpdateSessionTittle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid session ID", err.Error())
		return
	}

	var req UpdateSessionTittleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userID := c.GetInt("user_id")

	err = h.useCase.UpdateSessionTitle(ctx, sessionID, userID, req.Title)
	if err != nil {
		h.logger.Error("Error updating session title", "error", err)
		rest.RespondError(c, http.StatusInternalServerError, "Error updating session title", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session title updated successfully"})
}

func (h *ChatHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid session ID", err.Error())
		return
	}

	userID := c.GetInt("user_id")
	if err := h.useCase.DeleteSession(ctx, sessionID, userID); err != nil {
		h.logger.Error("Error deleting session", "error", err)
		rest.RespondError(c, http.StatusInternalServerError, "Error deleting session", "")
		return
	}

	h.logger.Info("Session deleted successfully", "session_id", sessionID, "user_id", userID)
	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) SendStream(c *gin.Context) {
	ctx := c.Request.Context()

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body")
		rest.RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userID := c.GetInt("user_id")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()

	answer, sources, sessionID, err := h.useCase.SendStream(ctx, userID, req.SessionID, req.Question,
		func(token string) error {
			c.SSEvent("chat_token", gin.H{"token": token})
			c.Writer.Flush()
			return nil
		})
	if err != nil {
		h.logger.Error("Failed to send stream message", "error", err)
		c.SSEvent("error", gin.H{"message": "Failed to send message"})
		c.Writer.Flush()
		return
	}

	c.SSEvent("chat_done", gin.H{
		"answer":     answer,
		"sources":    sources,
		"session_id": sessionID,
	})
	c.Writer.Flush()
}
