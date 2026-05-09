package postgres

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
)

type LLMUsageRepository struct {
	db *gorm.DB
}

func NewLLMUsageRepository(db *gorm.DB) *LLMUsageRepository {
	return &LLMUsageRepository{db: db}
}

func (r *LLMUsageRepository) Create(ctx context.Context, usage *domain.LLMUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}
