package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
)

type groupListResponse struct {
	Items []*domain.DocumentGroup `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
}

type DocumentGroupHandler struct {
	usecase *application.DocumentGroupUsecase
}

func NewDocumentGroupHandler(uc *application.DocumentGroupUsecase) *DocumentGroupHandler {
	return &DocumentGroupHandler{usecase: uc}
}

type createGroupRequest struct {
	Name           string `json:"name" binding:"required"`
	DocumentIDs    []int  `json:"document_ids"`
	AllowDownloads *bool  `json:"allow_downloads"`
}

type updateGroupRequest struct {
	Name           *string `json:"name,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
	AllowDownloads *bool   `json:"allow_downloads,omitempty"`
}

type addDocumentsRequest struct {
	DocumentIDs []int `json:"document_ids" binding:"required"`
}

func (h *DocumentGroupHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	allowDownloads := true
	if req.AllowDownloads != nil {
		allowDownloads = *req.AllowDownloads
	}

	userID := c.GetInt("user_id")
	group, err := h.usecase.Create(ctx, userID, req.Name, req.DocumentIDs, allowDownloads)
	if err != nil {
		if errors.Is(err, application.ErrInvalidDocument) {
			rest.RespondError(c, http.StatusBadRequest, err.Error(), "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Failed to create group", err.Error())
		return
	}

	c.JSON(http.StatusCreated, group)
}

func (h *DocumentGroupHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	params := parsePagination(c, 20)
	userID := c.GetInt("user_id")
	groups, total, err := h.usecase.List(ctx, userID, params.Page, params.Limit)
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Failed to list groups", err.Error())
		return
	}

	c.JSON(http.StatusOK, groupListResponse{
		Items: groups,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	})
}

func (h *DocumentGroupHandler) Get(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid group ID", "")
		return
	}

	userID := c.GetInt("user_id")
	group, err := h.usecase.Get(ctx, id, userID)
	if err != nil {
		if errors.Is(err, application.ErrGroupNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Group not found", "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Failed to get group", err.Error())
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *DocumentGroupHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid group ID", "")
		return
	}

	var req updateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userID := c.GetInt("user_id")
	if err := h.usecase.Update(ctx, id, userID, req.Name, req.IsActive, req.AllowDownloads); err != nil {
		if errors.Is(err, application.ErrGroupNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Group not found", "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Failed to update group", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *DocumentGroupHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid group ID", "")
		return
	}

	userID := c.GetInt("user_id")
	if err := h.usecase.Delete(ctx, id, userID); err != nil {
		if errors.Is(err, application.ErrGroupNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Group not found", "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Failed to delete group", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *DocumentGroupHandler) GetUsage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid group ID", "")
		return
	}

	userID := c.GetInt("user_id")
	summary, err := h.usecase.GetUsage(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, application.ErrGroupNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Group not found", "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Failed to get usage", err.Error())
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *DocumentGroupHandler) AddDocuments(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid group ID", "")
		return
	}

	var req addDocumentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	userID := c.GetInt("user_id")
	if err := h.usecase.AddDocuments(ctx, groupID, userID, req.DocumentIDs); err != nil {
		switch {
		case errors.Is(err, application.ErrGroupNotFound):
			rest.RespondError(c, http.StatusNotFound, "Group not found", "")
		case errors.Is(err, application.ErrInvalidDocument):
			rest.RespondError(c, http.StatusBadRequest, err.Error(), "")
		default:
			rest.RespondError(c, http.StatusInternalServerError, "Failed to add documents", err.Error())
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *DocumentGroupHandler) RemoveDocument(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid group ID", "")
		return
	}

	docID, err := strconv.Atoi(c.Param("doc_id"))
	if err != nil {
		rest.RespondError(c, http.StatusBadRequest, "Invalid document ID", "")
		return
	}

	userID := c.GetInt("user_id")
	if err := h.usecase.RemoveDocument(ctx, groupID, docID, userID); err != nil {
		if errors.Is(err, application.ErrGroupNotFound) {
			rest.RespondError(c, http.StatusNotFound, "Group not found", "")
			return
		}
		rest.RespondError(c, http.StatusInternalServerError, "Failed to remove document", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
