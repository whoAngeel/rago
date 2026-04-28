package ports

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindById(ctx context.Context, id int) (*domain.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) (*domain.Session, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	Revoke(ctx context.Context, refreshToken string) error
}
