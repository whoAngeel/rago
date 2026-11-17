package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tmc/langchaingo/schema"

	"github.com/whoAngeel/rago/internal/core/ports"
)

type contextKey string

// ForceOCRKey, when set to true in the context, sends every page through OCR.
const ForceOCRKey contextKey = "force_ocr"

const ocrDPI = "300"

// PDFParser extracts text per page with pdftotext, cleans it, and falls back
// to OCR for pages without a usable text layer.
type PDFParser struct {
	ocr ports.OCR
}

// NewPDFParser builds a PDF parser. ocr may be nil: pages that need OCR are
// then skipped with a warning.
func NewPDFParser(ocr ports.OCR) *PDFParser {
	return &PDFParser{ocr: ocr}
}

func (p *PDFParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	tmpDir, err := os.MkdirTemp("", "pdf-parse-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input.pdf")
	f, err := os.Create(inputPath)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	if _, err := io.Copy(f, reader); err != nil {
		f.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}

	rawPages, err := extractPages(ctx, inputPath)
	if err != nil {
		return nil, err
	}
	pageCount := len(rawPages)

	cleaned := make([]string, pageCount)
	for i, raw := range rawPages {
		cleaned[i] = cleanPageText(raw)
	}
	cleaned = removeRepeatedLines(cleaned)

	forceOCR, _ := ctx.Value(ForceOCRKey).(bool)

	var docs []schema.Document
	for i, text := range cleaned {
		pageNum := i + 1
		source := "text"
		if forceOCR || needsOCR(text) {
			text = p.ocrPage(ctx, inputPath, tmpDir, pageNum)
			source = "ocr"
		}
		if text == "" {
			continue
		}
		docs = append(docs, schema.Document{
			PageContent: text,
			Metadata: map[string]any{
				"page_number": pageNum,
				"page_count":  pageCount,
				"source":      source,
			},
		})
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no text extracted from PDF (%d pages)", pageCount)
	}
	return docs, nil
}

// extractPages runs pdftotext once and splits its output on form feeds; page
// N of the result is page N of the document.
func extractPages(ctx context.Context, pdfPath string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "pdftotext", "-enc", "UTF-8", pdfPath, "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftotext: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	pages := strings.Split(stdout.String(), "\f")
	// pdftotext terminates every page with a form feed, so the last element
	// is the empty tail after the final one.
	if n := len(pages); n > 0 && strings.TrimSpace(pages[n-1]) == "" {
		pages = pages[:n-1]
	}
	return pages, nil
}

// ocrPage rasterises one page and recognises it. It returns "" (after
// logging) when OCR is unavailable or fails, so the page is skipped.
func (p *PDFParser) ocrPage(ctx context.Context, pdfPath, tmpDir string, page int) string {
	if p.ocr == nil {
		slog.Warn("pdf: page needs OCR but no OCR engine is configured; skipping", "page", page)
		return ""
	}
	prefix := filepath.Join(tmpDir, "page-"+strconv.Itoa(page))
	n := strconv.Itoa(page)
	cmd := exec.CommandContext(ctx, "pdftoppm", "-png", "-r", ocrDPI, "-f", n, "-l", n, "-singlefile", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("pdf: pdftoppm failed; skipping page", "page", page, "err", err, "output", strings.TrimSpace(string(out)))
		return ""
	}
	pngPath := prefix + ".png"
	defer os.Remove(pngPath)
	png, err := os.ReadFile(pngPath)
	if err != nil {
		slog.Warn("pdf: reading rasterised page failed; skipping page", "page", page, "err", err)
		return ""
	}
	text, err := p.ocr.Recognize(ctx, png)
	if err != nil {
		slog.Warn("pdf: OCR failed; skipping page", "page", page, "err", err)
		return ""
	}
	text = cleanPageText(text)
	if text == "" {
		slog.Warn("pdf: OCR returned no text; skipping page", "page", page)
	}
	return text
}
