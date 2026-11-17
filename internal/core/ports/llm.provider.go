package ports

import "context"

type LLMResult struct {
	Answer       string
	Model        string
	InputTokens  int
	OutputTokens int
}

type LLMProvider interface {
	GenerateAnswer(ctx context.Context, system, prompt string) (LLMResult, error)
	Stream(ctx context.Context, system, prompt string, onToken func(token string) error) (LLMResult, error)
}
