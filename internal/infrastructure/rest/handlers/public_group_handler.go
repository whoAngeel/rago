package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
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
			rest.RespondError(c, http.StatusNotFound, "Grupo no encontrado", "")
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
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "group_inactive",
				"message": "El grupo está desactivado. Tu intento de chat fue registrado.",
			})
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

	isView := c.Query("view") == "true" || strings.Contains(c.Request.URL.Path, "/view")

	contentType := "application/octet-stream"
	disposition := "attachment"
	if isView {
		disposition = "inline"
		if len(filename) > 4 && filename[len(filename)-4:] == ".pdf" {
			contentType = "application/pdf"
		}
	}

	c.Header("Content-Disposition", disposition+"; filename=\""+filename+"\"")
	c.Header("Content-Type", contentType)
	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}
