package parser

import (
	"context"
	"fmt"
	"io"
)

type PlainTextAdapter struct {
}

func NewPlainTextAdapter() *PlainTextAdapter {
	return &PlainTextAdapter{}
}

func (p *PlainTextAdapter) Parse(ctx context.Context, reader io.Reader, contentType string) (string, error) {
	if contentType != "text/plain" {
		return "", fmt.Errorf("unsupported content type")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
