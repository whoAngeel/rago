package openrouter

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type OpenRouterAdapter struct {
	llm         llms.Model
	model       string
	temperature float64
	maxTokens   int
}

// NewOpenRouterAdapter builds the chat adapter. maxTokens <= 0 leaves the
// completion length to the provider.
func NewOpenRouterAdapter(orKey, baseUrl, model string, temperature float64, maxTokens int) (*OpenRouterAdapter, error) {
	llm, err := openai.New(
		openai.WithToken(orKey),
		openai.WithBaseURL(baseUrl),
		openai.WithModel(model),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating llm: %w", err)
	}
	return &OpenRouterAdapter{llm: llm, model: model, temperature: temperature, maxTokens: maxTokens}, nil
}

func (or *OpenRouterAdapter) callOptions(extra ...llms.CallOption) []llms.CallOption {
	opts := []llms.CallOption{llms.WithTemperature(or.temperature)}
	if or.maxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(or.maxTokens))
	}
	return append(opts, extra...)
}

func (or *OpenRouterAdapter) generate(ctx context.Context, system, prompt string, extra ...llms.CallOption) (ports.LLMResult, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}
	resp, err := or.llm.GenerateContent(ctx, messages, or.callOptions(extra...)...)
	if err != nil {
		return ports.LLMResult{}, err
	}
	if len(resp.Choices) == 0 {
		return ports.LLMResult{}, errors.New("llm returned no choices")
	}
	completion := resp.Choices[0].Content
	return ports.LLMResult{
		Answer:       completion,
		Model:        or.model,
		InputTokens:  (len([]rune(system)) + len([]rune(prompt))) / 4,
		OutputTokens: len([]rune(completion)) / 4,
	}, nil
}

func (or *OpenRouterAdapter) GenerateAnswer(ctx context.Context, system, prompt string) (ports.LLMResult, error) {
	res, err := or.generate(ctx, system, prompt)
	if err != nil {
		return ports.LLMResult{}, fmt.Errorf("error generating answer: %w", err)
	}
	return res, nil
}

func (or *OpenRouterAdapter) Stream(ctx context.Context, system, prompt string, onToken func(token string) error) (ports.LLMResult, error) {
	res, err := or.generate(ctx, system, prompt,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			return onToken(string(chunk))
		}),
	)
	if err != nil {
		return ports.LLMResult{}, fmt.Errorf("error streaming answer: %w", err)
	}
	return res, nil
}
