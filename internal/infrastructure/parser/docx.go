package parser

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"github.com/unidoc/unioffice/document"
)

type DOCXParser struct{}

func NewDOCXParser() *DOCXParser {
	return &DOCXParser{}
}

func (p *DOCXParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	tmpDir, err := os.MkdirTemp("", "docx-parse-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "input.docx")
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, reader); err != nil {
		f.Close()
		return nil, err
	}
	f.Close()

	doc, err := document.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	var paragraphs []string
	for _, p := range doc.Paragraphs() {
		var lines []string
		for _, r := range p.Runs() {
			lines = append(lines, r.Text())
		}
		if text := strings.TrimSpace(strings.Join(lines, "")); text != "" {
			paragraphs = append(paragraphs, text)
		}
	}

	content := strings.Join(paragraphs, "\n\n")
	return []schema.Document{{
		PageContent: content,
		Metadata: map[string]any{
			"content_type": contentType,
		},
	}}, nil
}
