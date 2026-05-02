package postgres

import (
	"context"
	"fmt"
	"time"

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

func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *domain.Document) (*domain.Document, error) {
	err := r.db.WithContext(ctx).Create(doc).Error
	return doc, err
}

func (r *DocumentRepository) FindDocumentByUserID(ctx context.Context, userID int) ([]*domain.Document, error) {
	var docs []*domain.Document
	err := r.db.WithContext(ctx).Where("user_id=?", userID).Order("created_at desc").Find(&docs).Error
	return docs, err
}

func (r *DocumentRepository) UpdateDocumentStatus(ctx context.Context, id int, status domain.DocumentStatus) error {
	return r.db.WithContext(ctx).Model(&domain.Document{}).Where("id = ?", id).Update("status", status).Error
}

func (r *DocumentRepository) UpdateDocument(ctx context.Context, doc *domain.Document) (*domain.Document, error) {
	err := r.db.WithContext(ctx).Model(&domain.Document{}).Where("id = ?", doc.ID).Updates(doc).Error
	return doc, err
}

func (r *DocumentRepository) FindByID(ctx context.Context, id int) (*domain.Document, error) {
	var doc domain.Document
	err := r.db.WithContext(ctx).First(&doc, id).Error
	return &doc, err
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, id int) error {
	document, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&document)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no document with ID [%d] found", id)
	}
	return nil
}

func (r *DocumentRepository) FindPendingDocuments(ctx context.Context, limit int) ([]*domain.Document, error) {
	var docs []*domain.Document
	err := r.db.WithContext(ctx).Where(
		"status = ? OR (status = ? AND updated_at < ?)",
		domain.StatusPending,
		domain.StatusProcessing,
		time.Now().Add(-5*time.Minute),
	).Order("created_at ASC").Limit(limit).Find(&docs).Error
	return docs, err
}

func (r *DocumentRepository) CreateProcessingStep(ctx context.Context, step *domain.ProcessingStep) error {
	return r.db.WithContext(ctx).Create(step).Error
}

func (r *DocumentRepository) UpdateProcessingStep(ctx context.Context, id, duration int, status, errMsg string) error {
	return r.db.WithContext(ctx).Model(&domain.ProcessingStep{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
		"duration_ms":   duration,
	}).Error
}
