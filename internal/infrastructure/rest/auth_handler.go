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
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
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
		RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err := h.usecase.Register(ctx, req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		h.logger.Error("Error register user", "error", err)
		RespondError(c, 400, "Error register user", err.Error())
		return
	}

	// h.logger.Debug("Result", "result", result)
	c.JSON(http.StatusAccepted, RegisterResponse{
		Message: "register successfully",
	})

}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,min=6"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body")
		RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	loginResult, err := h.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Error on login", "error", err)
		RespondError(c, http.StatusBadRequest, "Error login", err.Error())
		return
	}

	h.logger.Debug("LOGIN RESULT", "access", loginResult.AccessToken, "refresh", loginResult.RefreshToken)

	c.JSON(http.StatusOK, loginResult)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body")
		RespondError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	result, err := h.usecase.Refresh(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Error("Error refresh token", "error", err)
		RespondError(c, http.StatusInternalServerError, "Error refresh token", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, 400, "refresh_token required", err.Error())
		return
	}

	if err := h.usecase.Logout(ctx, req.RefreshToken); err != nil {
		h.logger.Error("Error logout", "error", err)
		RespondError(c, 500, "Error logout", err.Error())
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{Message: "logout successfully"})
}
