package handlers

import (
	"context"
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
}

func (h *DocumentHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userId := c.GetInt("user_id")

	documents, err := h.usecase.GetUsersDocuments(ctx, userId)
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Error getting documents", err.Error())
		return
	}

	h.logger.Debug("User id", "user_id", userId)

	c.JSON(http.StatusOK, ListResponse{
		Items: documents,
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
	allowedExts := map[string]bool{".txt": true, ".pdf": true, ".csv": true, ".json": true, ".docx": true, ".xlsx": true}
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
	doc, err := h.usecase.Upload(ctx, userId, file.Filename, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Upload failed", err.Error())
		return
	}

	h.logger.Info("File uploaded", "file", file.Filename)

	c.JSON(http.StatusCreated, UploadResponse{
		ID:       doc.ID,
		Filename: file.Filename,
		Status:   string(doc.Status),
	})
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
