package ports

import "context"

type Embedder interface {
	ComputeEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	EmbedText(ctx context.Context, text string) ([]float32, error)
}
