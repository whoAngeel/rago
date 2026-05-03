package ports

import "context"

type LLMProvider interface {
	GenerateAnswer(ctx context.Context, prompt string) (string, error)
	Stream(ctx context.Context, prompt string, onToken func(token string) error) (string, error)
}
