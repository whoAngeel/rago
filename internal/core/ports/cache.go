package ports

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// ParseCache stores parsed document pages by content hash and parser signature.
type ParseCache interface {
	Get(ctx context.Context, contentHash, signature string) ([]schema.Document, bool, error)
	// Put upserts the pages for the key.
	Put(ctx context.Context, contentHash, signature string, docs []schema.Document) error
}

// EmbeddingCache stores chunk embeddings by text hash, model and dimension.
type EmbeddingCache interface {
	Get(ctx context.Context, textHash, model string, dim int) ([]float32, bool, error)
	Put(ctx context.Context, textHash, model string, dim int, vector []float32) error
}
