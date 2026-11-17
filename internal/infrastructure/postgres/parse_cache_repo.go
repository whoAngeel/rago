package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ParseCacheRepository struct {
	db *gorm.DB
}

func NewParseCacheRepository(db *gorm.DB) *ParseCacheRepository {
	return &ParseCacheRepository{db: db}
}

type cachedPage struct {
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata"`
}

func (r *ParseCacheRepository) Get(ctx context.Context, contentHash, signature string) ([]schema.Document, bool, error) {
	var row domain.ParsedContent
	err := r.db.WithContext(ctx).
		Where("content_hash = ? AND parser_signature = ?", contentHash, signature).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var pages []cachedPage
	if err := json.Unmarshal([]byte(row.Pages), &pages); err != nil {
		return nil, false, err
	}
	docs := make([]schema.Document, len(pages))
	for i, p := range pages {
		docs[i] = schema.Document{PageContent: p.Text, Metadata: p.Metadata}
	}
	return docs, true, nil
}

func (r *ParseCacheRepository) Put(ctx context.Context, contentHash, signature string, docs []schema.Document) error {
	pages := make([]cachedPage, len(docs))
	for i, d := range docs {
		pages[i] = cachedPage{Text: d.PageContent, Metadata: d.Metadata}
	}
	raw, err := json.Marshal(pages)
	if err != nil {
		return err
	}
	row := domain.ParsedContent{ContentHash: contentHash, ParserSignature: signature, Pages: string(raw)}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "content_hash"}, {Name: "parser_signature"}},
		DoUpdates: clause.AssignmentColumns([]string{"pages"}),
	}).Create(&row).Error
}
