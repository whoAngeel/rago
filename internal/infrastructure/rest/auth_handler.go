package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type AuthHandler struct {
	usecase *application.AuthUsecase
	logger  ports.Logger
}

func NewAuthHandler(uc *application.AuthUsecase, log ports.Logger) *AuthHandler {
	return &AuthHandler{
		usecase: uc,
		logger:  log,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email,required"`
	Password string `json:"password,required"`
	Role     string `json:"role,default=viewer"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.usecase.Register(ctx, req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		h.logger.Error("Error register user", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error register user"})
		return
	}

	// h.logger.Debug("Result", "result", result)
	c.JSON(http.StatusAccepted, RegisterResponse{
		Message: "register successfully",
	})

}
