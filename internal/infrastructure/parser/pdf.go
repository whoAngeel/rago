package parser

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/tmc/langchaingo/schema"
)

type PDFParser struct{}

func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	tmpDir, err := os.MkdirTemp("", "pdf-parse-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input.pdf")
	f, err := os.Create(inputPath)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, reader); err != nil {
		f.Close()
		return nil, err
	}
	f.Close()

	pdfFile, pdfReader, err := pdf.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("opening pdf: %w", err)
	}
	defer pdfFile.Close()

	totalPages := pdfReader.NumPage()
	var pages []string

	for i := 1; i <= totalPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		pages = append(pages, text)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no readable text found in PDF")
	}

	var docs []schema.Document
	for i, pageText := range pages {
		docs = append(docs, schema.Document{
			PageContent: pageText,
			Metadata: map[string]any{
				"page_number": i + 1,
				"page_count":  len(pages),
			},
		})
	}
	return docs, nil
}
