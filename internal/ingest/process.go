package ingest

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type Processor struct {
	Splitter *textsplitter.RecursiveCharacter
}

func (p *Processor) PrepareDocuments(
	ctx context.Context,
	filename, content string,
) ([]schema.Document, error) {
	// crear documento base
	doc := schema.Document{
		PageContent: content,
		Metadata: map[string]any{
			"source": filename,
		},
	}

	// split recursivo
	chunks, err := p.Splitter.SplitDocuments([]schema.Document{doc})
	if err != nil {
		return nil, fmt.Errorf("error splitting document: %w", err)
	}
	return chunks, nil
}
