package postgres

import (
	"context"

	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{
		db: db,
	}
}
func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *domain.Document) error {
	result := r.db.WithContext(ctx).Create(doc)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *DocumentRepository) FindDocumentByUserID(ctx context.Context, userID int) ([]*domain.Document, error) {
	var docs []*domain.Document
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&docs).Error
	return docs, err
}

func (r *DocumentRepository) UpdateDocumentStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Document{}).Where("id = ?", id).Update("status", status).Error
}
