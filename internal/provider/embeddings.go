package provider

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

type Embedder struct {
	client *embeddings.EmbedderImpl
}

func NewEmbedder(apiKey, baseUrl, model string) (*Embedder, error) {
	llm, err := openai.New(openai.WithToken(apiKey), openai.WithBaseURL(baseUrl), openai.WithEmbeddingModel(model))
	if err != nil {
		return nil, fmt.Errorf("error creating openai client: %w", err)
	}
	embed, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("error initializing embedder: %w", err)
	}
	return &Embedder{client: embed}, nil

}

func (e *Embedder) ComputeEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	vectors, err := e.client.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("error computing embeddings: %w", err)
	}

	return vectors, nil
}
