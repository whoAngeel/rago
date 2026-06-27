package parser

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/tmc/langchaingo/schema"
)

type PDFParser struct {
	imageParser *ImageParser
}

func NewPDFParser() *PDFParser {
	return &PDFParser{imageParser: NewImageParser()}
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
		return p.ocrFallback(ctx, inputPath, totalPages)
	}

	var docs []schema.Document
	for i, pageText := range pages {
		docs = append(docs, schema.Document{
			PageContent: pageText,
			Metadata: map[string]any{
				"page_number": i + 1,
				"page_count":  totalPages,
			},
		})
	}
	return docs, nil
}

func (p *PDFParser) ocrFallback(ctx context.Context, pdfPath string, totalPages int) ([]schema.Document, error) {
	outDir, err := os.MkdirTemp("", "pdf-ocr-*")
	if err != nil {
		return nil, fmt.Errorf("create ocr temp dir: %w", err)
	}
	defer os.RemoveAll(outDir)

	outPrefix := filepath.Join(outDir, "page")
	cmd := exec.CommandContext(ctx, "pdftoppm", "-png", "-r", "200", pdfPath, outPrefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w: %s", err, string(out))
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		return nil, fmt.Errorf("read ocr dir: %w", err)
	}

	var docs []schema.Document
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		imgPath := filepath.Join(outDir, entry.Name())
		f, err := os.Open(imgPath)
		if err != nil {
			continue
		}
		pageDocs, err := p.imageParser.Parse(ctx, f, "image/png")
		f.Close()
		if err != nil {
			continue
		}
		for i := range pageDocs {
			pageDocs[i].Metadata["total_pages"] = totalPages
			pageDocs[i].Metadata["source"] = "ocr"
		}
		docs = append(docs, pageDocs...)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no text extracted from scanned PDF")
	}
	return docs, nil
}
