package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
)

type PublicGroupHandler struct {
	usecase *application.PublicChatUsecase
}

func NewPublicGroupHandler(uc *application.PublicChatUsecase) *PublicGroupHandler {
	return &PublicGroupHandler{usecase: uc}
}

func (h *PublicGroupHandler) GetGroupInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	slug := c.Param("slug")
	info, err := h.usecase.GetGroupInfo(ctx, slug)
	if err != nil {
		if errors.Is(err, application.ErrGroupInactive) {
			rest.RespondError(c, http.StatusNotFound, "Este chat no está disponible", "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Error getting group info", err.Error())
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *PublicGroupHandler) Chat(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	slug := c.Param("slug")

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	answer, err := h.usecase.Chat(ctx, slug, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrGroupInactive):
			rest.RespondError(c, http.StatusNotFound, "Este chat no está disponible", "")
		case errors.Is(err, application.ErrQuotaExceeded):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "quota_exceeded",
				"message": "Este chat ha alcanzado su límite de mensajes",
			})
		default:
			rest.RespondError(c, http.StatusInternalServerError, "Chat error", err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}

func (h *PublicGroupHandler) DownloadDocument(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	slug := c.Param("slug")
	docID, err := strconv.Atoi(c.Param("doc_id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid document ID", "")
		return
	}

	reader, filename, err := h.usecase.DownloadDocument(ctx, slug, docID)
	if err != nil {
		if errors.Is(err, application.ErrGroupInactive) {
			rest.RespondError(c, http.StatusNotFound, "Este chat no está disponible", "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Download error", err.Error())
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "application/octet-stream")
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}
