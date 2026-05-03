package postgres

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
)

type SystemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

func (r *SystemConfigRepository) Get(ctx context.Context, key string) (string, error) {
	var config domain.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
	return config.Value, err
}

func (r *SystemConfigRepository) Set(ctx context.Context, key, value string) error {
	err := r.db.WithContext(ctx).
		Where("key = ?", key).
		Assign(domain.SystemConfig{Value: value}).
		FirstOrCreate(&domain.SystemConfig{}).
		Error
	return err
}
