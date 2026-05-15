package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DocumentGroupRepository struct {
	db *gorm.DB
}

func NewDocumentGroupRepository(db *gorm.DB) *DocumentGroupRepository {
	return &DocumentGroupRepository{db: db}
}

func (r *DocumentGroupRepository) Create(ctx context.Context, group *domain.DocumentGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *DocumentGroupRepository) FindByID(ctx context.Context, id, userID int) (*domain.DocumentGroup, error) {
	var group domain.DocumentGroup
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ids, err := r.FindDocumentIDs(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	group.DocumentIDs = ids
	return &group, nil
}

func (r *DocumentGroupRepository) FindByUserID(ctx context.Context, userID, page, limit int) ([]*domain.DocumentGroup, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&domain.DocumentGroup{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var groups []*domain.DocumentGroup
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}
	for _, g := range groups {
		ids, err := r.FindDocumentIDs(ctx, g.ID)
		if err != nil {
			return nil, 0, err
		}
		g.DocumentIDs = ids
	}
	return groups, total, nil
}

func (r *DocumentGroupRepository) Update(ctx context.Context, group *domain.DocumentGroup) error {
	return r.db.WithContext(ctx).
		Model(&domain.DocumentGroup{}).
		Where("id = ? AND user_id = ?", group.ID, group.UserID).
		Updates(map[string]interface{}{
			"name":            group.Name,
			"is_active":       group.IsActive,
			"allow_downloads": group.AllowDownloads,
		}).Error
}

func (r *DocumentGroupRepository) Delete(ctx context.Context, id, userID int) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.DocumentGroup{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("group %d not found", id)
	}
	return nil
}

func (r *DocumentGroupRepository) AddDocuments(ctx context.Context, groupID int, documentIDs []int) error {
	items := make([]domain.DocumentGroupItem, 0, len(documentIDs))
	for _, docID := range documentIDs {
		items = append(items, domain.DocumentGroupItem{GroupID: groupID, DocumentID: docID})
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&items).Error
}

func (r *DocumentGroupRepository) RemoveDocument(ctx context.Context, groupID, documentID int) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND document_id = ?", groupID, documentID).
		Delete(&domain.DocumentGroupItem{}).Error
}

func (r *DocumentGroupRepository) FindDocumentIDs(ctx context.Context, groupID int) ([]int, error) {
	var items []domain.DocumentGroupItem
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.DocumentID)
	}
	return ids, nil
}

func (r *DocumentGroupRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.DocumentGroup{}).
		Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *DocumentGroupRepository) FindBySlug(ctx context.Context, slug string) (*domain.DocumentGroup, error) {
	var group domain.DocumentGroup
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ids, err := r.FindDocumentIDs(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	group.DocumentIDs = ids
	return &group, nil
}

func (r *DocumentGroupRepository) IncrementAttempts(ctx context.Context, groupID int) error {
	return r.db.WithContext(ctx).Model(&domain.DocumentGroup{}).
		Where("id = ?", groupID).
		UpdateColumn("chat_attempts", gorm.Expr("chat_attempts + 1")).Error
}

func (r *DocumentGroupRepository) IncrementQuotaUsed(ctx context.Context, groupID int) error {
	return r.db.WithContext(ctx).Model(&domain.DocumentGroup{}).
		Where("id = ?", groupID).
		UpdateColumn("chat_quota_used", gorm.Expr("chat_quota_used + 1")).Error
}
