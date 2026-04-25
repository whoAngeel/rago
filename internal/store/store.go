package store

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

type VectorStore interface {
	createCollection(ctx context.Context, name string, size int) error
	upsertDocuments(ctx context.Context, collection string, docs []schema.Document, vector [][]float32) error
}
