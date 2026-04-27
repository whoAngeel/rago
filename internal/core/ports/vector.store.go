package ports

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

type VectorStore interface {
	CreateCollection(ctx context.Context, name string, size int) error
	UpsertDocuments(ctx context.Context, collection string, docs []schema.Document, vectors [][]float32) error
	Search(ctx context.Context, queryVector []float32, limit int) (SearchResult, error)
	GetPointsCount(ctx context.Context, collection string) (int64, error)
	DeleteColletion(ctx context.Context, collection string) error
}

type SearchResult struct {
	Document schema.Document
	score    float32
}
