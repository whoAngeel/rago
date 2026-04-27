package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

type EmbedderAdapter struct {
	client *embeddings.EmbedderImpl
}

func NewEmbedderAdapter(apiKey, baseUrl, model string) (*EmbedderAdapter, error) {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithBaseURL(baseUrl),
		openai.WithEmbeddingModel(model),
		openai.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating openai client: %w", err)
	}
	embed, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("error initializing embedder: %w", err)
	}
	return &EmbedderAdapter{client: embed}, nil
}

func (e *EmbedderAdapter) ComputeEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	vectors, err := e.client.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("error computing embeddings: %w", err)
	}

	return vectors, nil
}

func (e *EmbedderAdapter) EmbedText(ctx context.Context, text string) ([]float32, error) {
	vectors, err := e.client.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("error embedding text: %w", err)
	}
	return vectors[0], nil
}
