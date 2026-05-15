package application

import (
	"context"

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

type UserUsecase struct {
	UserRepo ports.UserRepository
	DocRepo  ports.DocumentRepository
}

func NewUserUsecase(userRepo ports.UserRepository, docRepo ports.DocumentRepository) *UserUsecase {
	return &UserUsecase{UserRepo: userRepo, DocRepo: docRepo}
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
