package openrouter

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type OpenRouterAdapter struct {
	llm   llms.Model
	model string
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
	return &OpenRouterAdapter{llm: llm, model: model}, nil
}

func (or *OpenRouterAdapter) GenerateAnswer(ctx context.Context, prompt string) (ports.LLMResult, error) {
	completion, err := llms.GenerateFromSinglePrompt(ctx, or.llm, prompt)
	if err != nil {
		return ports.LLMResult{}, fmt.Errorf("error generating answer: %w", err)
	}
	return ports.LLMResult{
		Answer:       completion,
		Model:        or.model,
		InputTokens:  len([]rune(prompt)) / 4,
		OutputTokens: len([]rune(completion)) / 4,
	}, nil
}

func (or *OpenRouterAdapter) Stream(ctx context.Context, prompt string, onToken func(token string) error) (ports.LLMResult, error) {
	completion, err := llms.GenerateFromSinglePrompt(ctx, or.llm, prompt,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			return onToken(string(chunk))
		}),
	)
	if err != nil {
		return ports.LLMResult{}, fmt.Errorf("error streaming answer: %w", err)
	}
	return ports.LLMResult{
		Answer:       completion,
		Model:        or.model,
		InputTokens:  len([]rune(prompt)) / 4,
		OutputTokens: len([]rune(completion)) / 4,
	}, nil
}
