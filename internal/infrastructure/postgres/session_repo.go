package postgres

import (
	"context"
	"time"

	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(
	db *gorm.DB,
) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	result := r.db.WithContext(ctx).Create(session)
	return session, result.Error
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.WithContext(ctx).Where("refresh_token = ? ", refreshToken).First(&session).Error
	return &session, err
}

func (r *SessionRepository) Revoke(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&domain.Session{}).
		Where("refresh_token = ?", token).
		Update("revoked_at", time.Now()).Error
}
