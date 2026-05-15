package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type ConfigHandler struct {
	configRepo ports.SystemConfigRepository
	logger     ports.Logger
}

func NewConfigHandler(repo ports.SystemConfigRepository, log ports.Logger) *ConfigHandler {
	return &ConfigHandler{
		configRepo: repo,
		logger:     log,
	}
}

func (h *ConfigHandler) GetSystemPrompt(c *gin.Context) {
	value, err := h.configRepo.Get(c.Request.Context(), "system_prompt")
	if err != nil {
		h.logger.Error("failed to get system prompt", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "system prompt not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"value": value})
}

type UpdateSystemPromptRequest struct {
	Value string `json:"value" binding:"required"`
}

func (h *ConfigHandler) UpdateSystemPrompt(c *gin.Context) {
	var req UpdateSystemPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configRepo.Set(c.Request.Context(), "system_prompt", req.Value); err != nil {
		h.logger.Error("failed to update system prompt", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update system prompt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": req.Value})
}
