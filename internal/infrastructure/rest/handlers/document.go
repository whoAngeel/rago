package handlers

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
)

type DocumentHandler struct {
	usecase *application.IngestDocumentUsecase
	logger  ports.Logger
	config  config.Config
}

func NewDocumentHandler(uc *application.IngestDocumentUsecase, log ports.Logger, config config.Config) *DocumentHandler {
	return &DocumentHandler{
		usecase: uc,
		logger:  log,
		config:  config,
	}
}

type ListResponse struct {
	Items []*domain.Document `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

type PaginationParams struct {
	Page  int
	Limit int
}

func parsePagination(c *gin.Context, defaultLimit int) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = defaultLimit
	}
	return PaginationParams{Page: page, Limit: limit}
}

func (h *DocumentHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userId := c.GetInt("user_id")

	p := parsePagination(c, 10)

	documents, total, err := h.usecase.GetUsersDocuments(ctx, userId, p.Page, p.Limit)
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Error getting documents", err.Error())
		return
	}

	h.logger.Debug("User id", "user_id", userId)

	c.JSON(http.StatusOK, ListResponse{
		Items: documents,
		Total: total,
		Page:  p.Page,
		Limit: p.Limit,
	})

}

type UploadResponse struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	file, err := c.FormFile("file")
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Error retrieving file", err.Error())
		return
	}
	if file.Size > h.config.MaxUploadSize {
		rest.RespondError(c, 413, "File too large", "")
		return
	}
	allowedExts := map[string]bool{".txt": true, ".pdf": true, ".csv": true, ".json": true, ".docx": true, ".xlsx": true, ".md": true}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		rest.RespondError(c, 400, "File type not allowed", "")
		return
	}
	userId := c.GetInt("user_id")
	src, err := file.Open()
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Error opening file", err.Error())
		return
	}
	defer src.Close()
	contentType := file.Header.Get("Content-Type")
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
		contentType = mediaType
	}
	doc, err := h.usecase.Upload(ctx, userId, file.Filename, src, file.Size, contentType)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrDuplicateDocument):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "duplicate_document",
				"message": "Ya tenés este archivo subido",
			})
		case errors.Is(err, application.ErrDocumentLimitReached):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "quota_exceeded",
				"message": "Alcanzaste el límite de documentos",
			})
		default:
			rest.RespondError(c, http.StatusInternalServerError, "Upload failed", err.Error())
		}
		return
	}

	h.logger.Info("File uploaded", "file", file.Filename)

	c.JSON(http.StatusCreated, UploadResponse{
		ID:       doc.ID,
		Filename: file.Filename,
		Status:   string(doc.Status),
	})
}

func (h *DocumentHandler) Steps(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid document ID", err.Error())
		return
	}

	userID := c.GetInt("user_id")

	steps, err := h.usecase.GetDocumentSteps(ctx, id, userID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Document not found", "")
		} else {
			rest.RespondError(c, http.StatusInternalServerError, "Error getting steps", err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, steps)
}

func (h *DocumentHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid document ID", err.Error())
		return
	}

	userID := c.GetInt("user_id")

	health, err := h.usecase.GetDocumentHealth(ctx, id, userID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Document not found", "")
		} else {
			rest.RespondError(c, http.StatusInternalServerError, "Error getting document health", err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, health)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, 400, "Invalid document ID", err.Error())
		return
	}

	if err := h.usecase.DeleteDocument(ctx, id); err != nil {
		rest.RespondError(c, 500, "Delete failed", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type SelectItem struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
}

func (h *DocumentHandler) ListSelect(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userId := c.GetInt("user_id")
	docs, err := h.usecase.GetDocumentsForSelect(ctx, userId)
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Error", err.Error())
		return
	}

	items := make([]SelectItem, 0, len(docs))
	for _, d := range docs {
		if d.Status == domain.StatusFailed {
			continue
		}
		items = append(items, SelectItem{ID: d.ID, Filename: d.Filename})
	}

	c.JSON(http.StatusOK, items)
}

func (h *DocumentHandler) Reprocess(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid document ID", err.Error())
		return
	}

	var req struct {
		ForceOCR bool `json:"force_ocr"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.ForceOCR = false
	}

	userID := c.GetInt("user_id")
	if err := h.usecase.ReprocessDocument(ctx, id, userID, req.ForceOCR); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Document not found", "")
		} else {
			rest.RespondError(c, http.StatusBadRequest, "Reprocess failed", err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
