package ports

import "context"

type LLMProvider interface {
	GenerateAnswer(ctx context.Context, prompt string) (string, error)
}
