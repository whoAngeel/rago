package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/infrastructure/rest"
)

type UserHandler struct {
	usecase *application.UserUsecase
}

func NewUserHandler(uc *application.UserUsecase) *UserHandler {
	return &UserHandler{usecase: uc}
}

func (h *UserHandler) Me(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userID := c.GetInt("user_id")

	profile, err := h.usecase.GetProfile(ctx, userID)
	if err != nil {
		rest.RespondError(c, http.StatusInternalServerError, "Error fetching profile", err.Error())
		return
	}

	c.JSON(http.StatusOK, profile)
}
