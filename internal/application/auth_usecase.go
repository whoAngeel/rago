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

type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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

func (au *AuthUsecase) Register(ctx context.Context, name, email, password, role string) error {
	existingUser, _ := au.UserRepository.FindByEmail(ctx, email)
	if existingUser != nil {
		return fmt.Errorf("email already register")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password")
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
	_, err = au.UserRepository.Create(ctx, user)

	return err
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

	role := resolveRoleName(existingUser.RoleID)

	access_token, err := auth.GenerateAccessToken(existingUser.ID, role, au.JWTSecret, au.AccessTokenExpiration)
	if err != nil {
		return nil, err
	}

	refresh_token, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if _, err := au.SessionRepository.Create(ctx, &domain.Session{
		UserID:       existingUser.ID,
		RefreshToken: refresh_token,
		AccessToken:  access_token,
		ExpiresAt:    time.Now().Add(au.RefreshTokenExpiration),
	}); err != nil {
		au.logger.Error("failed to save session", "error", err)
	}

	return &LoginResult{
		AccessToken:  access_token,
		RefreshToken: refresh_token,
	}, nil
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

	role := resolveRoleName(user.RoleID)
	accessToken, err := auth.GenerateAccessToken(user.ID, role, au.JWTSecret, au.AccessTokenExpiration)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: token,
	}, nil
}

func (au *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	return au.SessionRepository.Revoke(ctx, refreshToken)
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
