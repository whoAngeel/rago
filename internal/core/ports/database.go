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

type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *domain.Document) (*domain.Document, error)
	UpdateDocument(ctx context.Context, doc *domain.Document) (*domain.Document, error)
	FindDocumentByUserID(ctx context.Context, userID int) ([]*domain.Document, error)
	UpdateDocumentStatus(ctx context.Context, id int, status domain.DocumentStatus) error
	FindByID(ctx context.Context, id int) (*domain.Document, error)
	DeleteDocument(ctx context.Context, id int) error
	FindPendingDocuments(ctx context.Context, limit int) ([]*domain.Document, error)
	CreateProcessingStep(ctx context.Context, step *domain.ProcessingStep) error
	UpdateProcessingStep(ctx context.Context, id, duration int, status, errMsg string) error
}

type ChatRepository interface {
	CreateSession(ctx context.Context, session *domain.ChatSession) error
	GetSession(ctx context.Context, id, userID int) (*domain.ChatSession, error)
	GetUserSessions(ctx context.Context, userID int) ([]*domain.ChatSession, error)
	UpdateSessionTitle(ctx context.Context, id, userID int, title string) error
	DeleteSession(ctx context.Context, id, userID int) error

	CreateMessage(ctx context.Context, msg *domain.ChatMessage) error
	GetMessages(ctx context.Context, sessionID, limit int) ([]*domain.ChatMessage, error)
}

type SystemConfigRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}
