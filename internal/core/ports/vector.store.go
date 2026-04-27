package ports

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

type VectorStore interface {
	CreateCollection(ctx context.Context, name string, size int) error
	UpsertDocuments(ctx context.Context, collection string, docs []schema.Document, vectors [][]float32) error
	Search(ctx context.Context, collection string, queryVector []float32, limit int) ([]SearchResult, error)
	GetPointsCount(ctx context.Context, collection string) (uint64, error)
	DeleteCollection(ctx context.Context, collection string) error
}

type SearchResult struct {
	Document schema.Document
	Score    float32
}
