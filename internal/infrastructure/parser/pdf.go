package parser

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/tmc/langchaingo/schema"
)

type contextKey string

const ForceOCRKey contextKey = "force_ocr"

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

	forceOCR, _ := ctx.Value(ForceOCRKey).(bool)
	if forceOCR {
		return p.ocrFallback(ctx, inputPath, totalPages)
	}

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
	var docs []schema.Document

	for page := 1; page <= totalPages; page++ {
		pageDir, err := os.MkdirTemp("", "pdf-ocr-page-*")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}

		outPrefix := filepath.Join(pageDir, "p")
		cmd := exec.CommandContext(ctx, "pdftoppm",
			"-png", "-r", "150",
			"-f", fmt.Sprintf("%d", page),
			"-l", fmt.Sprintf("%d", page),
			pdfPath, outPrefix,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			slog.Warn("pdf ocr: pdftoppm failed", "page", page, "total", totalPages, "err", err, "output", string(out))
			os.RemoveAll(pageDir)
			continue
		}

		entries, _ := os.ReadDir(pageDir)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			imgPath := filepath.Join(pageDir, entry.Name())
			f, err := os.Open(imgPath)
			if err != nil {
				continue
			}
			pageDocs, err := p.imageParser.Parse(ctx, f, "image/png")
			f.Close()
			if err != nil {
				slog.Warn("pdf ocr: tesseract failed", "page", page, "total", totalPages, "err", err)
				continue
			}
			for i := range pageDocs {
				pageDocs[i].Metadata["page_number"] = page
				pageDocs[i].Metadata["total_pages"] = totalPages
				pageDocs[i].Metadata["source"] = "ocr"
			}
			docs = append(docs, pageDocs...)
		}

		os.RemoveAll(pageDir)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no text extracted from scanned PDF (%d pages attempted)", totalPages)
	}
	if len(docs) < totalPages {
		slog.Warn("pdf ocr: partial extraction", "extracted_pages", len(docs), "total_pages", totalPages)
	}
	return docs, nil
}
