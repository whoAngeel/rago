package postgres

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
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

func (r *LLMUsageRepository) SumByGroupID(ctx context.Context, groupID int) (ports.GroupUsageSummary, error) {
	var result ports.GroupUsageSummary
	err := r.db.WithContext(ctx).
		Model(&domain.LLMUsage{}).
		Select("COUNT(*) as total_calls, COALESCE(SUM(input_tokens), 0) as input_tokens, COALESCE(SUM(output_tokens), 0) as output_tokens").
		Where("group_id = ?", groupID).
		Scan(&result).Error
	return result, err
}
