package application

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type UserProfile struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	MaxDocuments  int    `json:"max_documents"`
	DocumentCount int64  `json:"document_count"`
	ChatQuota     int    `json:"chat_quota"`
	ChatQuotaUsed int    `json:"chat_quota_used"`
}

type DocStatusCount struct {
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	Pending   int64 `json:"pending"`
}

type DashboardStats struct {
	TotalDocuments int64          `json:"total_documents"`
	ByStatus       DocStatusCount `json:"by_status"`
	StorageUsed    int64          `json:"storage_used"`
	ChatSessions   int64          `json:"chat_sessions"`
	TotalGroups    int64          `json:"total_groups"`
}

type UserUsecase struct {
	UserRepo  ports.UserRepository
	DocRepo   ports.DocumentRepository
	ChatRepo  ports.ChatRepository
	GroupRepo ports.DocumentGroupRepository
}

func NewUserUsecase(userRepo ports.UserRepository, docRepo ports.DocumentRepository, chatRepo ports.ChatRepository, groupRepo ports.DocumentGroupRepository) *UserUsecase {
	return &UserUsecase{UserRepo: userRepo, DocRepo: docRepo, ChatRepo: chatRepo, GroupRepo: groupRepo}
}

func (uc *UserUsecase) GetProfile(ctx context.Context, userID int) (*UserProfile, error) {
	user, err := uc.UserRepo.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}

	count, err := uc.DocRepo.CountDocumentsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserProfile{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Role:          resolveRoleName(user.RoleID),
		MaxDocuments:  user.MaxDocuments,
		DocumentCount: count,
		ChatQuota:     user.ChatQuota,
		ChatQuotaUsed: user.ChatQuotaUsed,
	}, nil
}

func (uc *UserUsecase) GetDashboardStats(ctx context.Context, userID int) (*DashboardStats, error) {
	total, err := uc.DocRepo.CountDocumentsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	completed, err := uc.DocRepo.CountByStatus(ctx, userID, domain.StatusCompleted)
	if err != nil {
		return nil, err
	}
	failed, err := uc.DocRepo.CountByStatus(ctx, userID, domain.StatusFailed)
	if err != nil {
		return nil, err
	}
	pending, err := uc.DocRepo.CountByStatus(ctx, userID, domain.StatusPending)
	if err != nil {
		return nil, err
	}

	storage, err := uc.DocRepo.SumSizeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var chatSessions int64
	if uc.ChatRepo != nil {
		sessions, err := uc.ChatRepo.GetUserSessions(ctx, userID)
		if err == nil {
			chatSessions = int64(len(sessions))
		}
	}

	var totalGroups int64
	if uc.GroupRepo != nil {
		_, total, err := uc.GroupRepo.FindByUserID(ctx, userID, 1, 100)
		if err == nil {
			totalGroups = total
		}
	}

	return &DashboardStats{
		TotalDocuments: total,
		ByStatus: DocStatusCount{
			Completed: completed,
			Failed:    failed,
			Pending:   pending,
		},
		StorageUsed:  storage,
		ChatSessions: chatSessions,
		TotalGroups:  totalGroups,
	}, nil
}
