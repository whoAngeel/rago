package postgres

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateSession(ctx context.Context, session *domain.ChatSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *ChatRepository) GetSession(ctx context.Context, id, userID int) (*domain.ChatSession, error) {
	var session domain.ChatSession
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&session).Error
	return &session, err
}

func (r *ChatRepository) GetUserSessions(ctx context.Context, userID int) ([]*domain.ChatSession, error) {
	var sessions []*domain.ChatSession
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}

func (r *ChatRepository) UpdateSessionTitle(ctx context.Context, id, userID int, title string) error {
	return r.db.WithContext(ctx).Model(&domain.ChatSession{}).Where("id = ? AND user_id = ?", id, userID).Update("title", title).Error
}

func (r *ChatRepository) DeleteSession(ctx context.Context, id, userID int) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&domain.ChatSession{}).Error
}

func (r *ChatRepository) CreateMessage(ctx context.Context, msg *domain.ChatMessage) error {
	err := r.db.WithContext(ctx).Create(msg).Error
	return err
}

func (r *ChatRepository) GetMessages(ctx context.Context, sessionID, limit int) ([]*domain.ChatMessage, error) {
	var messages []*domain.ChatMessage
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at desc").Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) GetAllMessages(ctx context.Context, sessionID int) ([]*domain.ChatMessage, error) {
	var messages []*domain.ChatMessage
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error
	return messages, err
}
