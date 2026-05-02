package parser

import (
	"context"
	"fmt"
	"io"

	"github.com/tmc/langchaingo/schema"
)

type PlainTextAdapter struct {
}

func NewPlainTextAdapter() *PlainTextAdapter {
	return &PlainTextAdapter{}
}

func (p *PlainTextAdapter) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	if contentType != "text/plain" {
		return nil, fmt.Errorf("unsupported content type")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return []schema.Document{{
		PageContent: string(data),
		Metadata:    map[string]any{"content_type": contentType},
	}}, nil
}
