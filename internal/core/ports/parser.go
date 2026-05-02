package ports

import (
	"context"
	"io"

	"github.com/tmc/langchaingo/schema"
)

type Parser interface {
	Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error)
}

type ParserRegistry interface {
	Resolve(contentType string) (Parser, error)
}
