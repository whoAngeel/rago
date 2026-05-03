package openrouter

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenRouterAdapter struct {
	llm llms.Model
}

func NewOpenRouterAdapter(orKey, baseUrl, model string) (*OpenRouterAdapter, error) {
	llm, err := openai.New(
		openai.WithToken(orKey),
		openai.WithBaseURL(baseUrl),
		openai.WithModel(model),
	)

	if err != nil {
		return nil, fmt.Errorf("error creating llm: %w", err)
	}
	return &OpenRouterAdapter{
		llm: llm,
	}, nil
}

func (or *OpenRouterAdapter) GenerateAnswer(ctx context.Context, prompt string) (string, error) {
	completion, err := llms.GenerateFromSinglePrompt(ctx, or.llm, prompt)
	if err != nil {
		return "", fmt.Errorf("error generating answer: %w", err)
	}
	return completion, nil
}

func (or *OpenRouterAdapter) Stream(ctx context.Context, prompt string, onToken func(token string) error) (string, error) {
	completion, err := llms.GenerateFromSinglePrompt(ctx, or.llm, prompt,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			return onToken(string(chunk))
		}),
	)
	if err != nil {
		return "", fmt.Errorf("error streaming answer: %w", err)
	}
	return completion, nil
}
