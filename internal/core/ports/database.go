package ports

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindById(ctx context.Context, id int) (*domain.User, error)
	IncrementChatQuotaUsed(ctx context.Context, userID int) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) (*domain.Session, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	Revoke(ctx context.Context, refreshToken string) error
}

type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *domain.Document) (*domain.Document, error)
	UpdateDocument(ctx context.Context, doc *domain.Document) (*domain.Document, error)
	FindDocumentByUserID(ctx context.Context, userID int, page, limit int) ([]*domain.Document, int64, error)
	UpdateDocumentStatus(ctx context.Context, id int, status domain.DocumentStatus) error
	FindByID(ctx context.Context, id int) (*domain.Document, error)
	DeleteDocument(ctx context.Context, id int) error
	FindPendingDocuments(ctx context.Context, limit int) ([]*domain.Document, error)
	CreateProcessingStep(ctx context.Context, step *domain.ProcessingStep) error
	UpdateProcessingStep(ctx context.Context, id, duration int, status, errMsg string) error
	FindStepsByDocumentID(ctx context.Context, docID int) ([]*domain.ProcessingStep, error)
	FindByChecksum(ctx context.Context, userID int, checksum string) (*domain.Document, error)
	CountDocumentsByUserID(ctx context.Context, userID int) (int64, error)
	FindDocumentsForSelect(ctx context.Context, userID int) ([]*domain.Document, error)
}

type ChatRepository interface {
	CreateSession(ctx context.Context, session *domain.ChatSession) error
	GetSession(ctx context.Context, id, userID int) (*domain.ChatSession, error)
	GetUserSessions(ctx context.Context, userID int) ([]*domain.ChatSession, error)
	UpdateSessionTitle(ctx context.Context, id, userID int, title string) error
	DeleteSession(ctx context.Context, id, userID int) error

	CreateMessage(ctx context.Context, msg *domain.ChatMessage) error
	GetMessages(ctx context.Context, sessionID, limit int) ([]*domain.ChatMessage, error)
	GetAllMessages(ctx context.Context, sessionID int) ([]*domain.ChatMessage, error)
}

type SystemConfigRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

type GroupUsageSummary struct {
	TotalCalls   int64 `json:"total_calls"`
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

type LLMUsageRepository interface {
	Create(ctx context.Context, usage *domain.LLMUsage) error
	SumByGroupID(ctx context.Context, groupID int) (GroupUsageSummary, error)
}

type DocumentGroupRepository interface {
	Create(ctx context.Context, group *domain.DocumentGroup) error
	FindByID(ctx context.Context, id, userID int) (*domain.DocumentGroup, error)
	FindBySlug(ctx context.Context, slug string) (*domain.DocumentGroup, error)
	FindByUserID(ctx context.Context, userID, page, limit int) ([]*domain.DocumentGroup, int64, error)
	Update(ctx context.Context, group *domain.DocumentGroup) error
	Delete(ctx context.Context, id, userID int) error
	AddDocuments(ctx context.Context, groupID int, documentIDs []int) error
	RemoveDocument(ctx context.Context, groupID, documentID int) error
	FindDocumentIDs(ctx context.Context, groupID int) ([]int, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	IncrementAttempts(ctx context.Context, groupID int) error
	IncrementQuotaUsed(ctx context.Context, groupID int) error
}
