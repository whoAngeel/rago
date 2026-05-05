package application

import (
	"context"
	"fmt"
	"time"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/auth"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	UserRepository         ports.UserRepository
	SessionRepository      ports.SessionRepository
	JWTSecret              string
	logger                 ports.Logger
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
}

type UserInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResult struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

func NewAuthUseCase(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	secret string,
	logger ports.Logger,
	atExpiration time.Duration,
	rtExpiration time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		UserRepository:         userRepo,
		SessionRepository:      sessionRepo,
		JWTSecret:              secret,
		AccessTokenExpiration:  atExpiration,
		RefreshTokenExpiration: rtExpiration,
		logger:                 logger,
	}
}

func (au *AuthUsecase) Register(ctx context.Context, name, email, password, role string) (*LoginResult, error) {
	existingUser, _ := au.UserRepository.FindByEmail(ctx, email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already register")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password")
	}

	roleID := 3 // default viewer
	switch role {
	case string(domain.RoleAdmin):
		roleID = 1
	case string(domain.RoleEditor):
		roleID = 2
	}
	user := &domain.User{
		Email:    email,
		Password: string(hash),
		Name:     name,
		RoleID:   roleID,
	}
	createdUser, err := au.UserRepository.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return au.generateAuthResult(ctx, createdUser)
}

func (au *AuthUsecase) Login(ctx context.Context, email, pass string) (*LoginResult, error) {
	existingUser, _ := au.UserRepository.FindByEmail(ctx, email)
	if existingUser == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(pass))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return au.generateAuthResult(ctx, existingUser)
}

func (au *AuthUsecase) Refresh(ctx context.Context, token string) (*LoginResult, error) {
	session, err := au.SessionRepository.FindByRefreshToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if session.RevokedAt != nil {
		return nil, fmt.Errorf("session revoked")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	user, err := au.UserRepository.FindById(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := au.SessionRepository.Revoke(ctx, token); err != nil {
		return nil, fmt.Errorf("revoking old session: %w", err)
	}

	return au.generateAuthResult(ctx, user)
}

func (au *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	return au.SessionRepository.Revoke(ctx, refreshToken)
}

func (au *AuthUsecase) generateAuthResult(ctx context.Context, user *domain.User) (*LoginResult, error) {
	roleName := resolveRoleName(user.RoleID)

	accessToken, err := auth.GenerateAccessToken(user.ID, roleName, au.JWTSecret, au.AccessTokenExpiration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if _, err := au.SessionRepository.Create(ctx, &domain.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		ExpiresAt:    time.Now().Add(au.RefreshTokenExpiration),
	}); err != nil {
		au.logger.Error("failed to save session", "error", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  roleName,
		},
	}, nil
}

func resolveRoleName(role int) string {
	switch role {
	case 1:
		return "admin"
	case 2:
		return "editor"
	case 3:
		return "viewer"
	default:
		return ""
	}
}
