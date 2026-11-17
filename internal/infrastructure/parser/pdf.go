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
	"sync"

	"github.com/tmc/langchaingo/schema"

	"github.com/whoAngeel/rago/internal/core/ports"
)

type contextKey string

// ForceOCRKey, when set to true in the context, sends every page through OCR.
const ForceOCRKey contextKey = "force_ocr"

const ocrDPI = "300"

// maxConcurrentOCR bounds the OCR calls in flight for one document.
const maxConcurrentOCR = 4

// PDFParser extracts text per page with pdftotext, cleans it, and falls back
// to OCR for pages without a usable text layer.
type PDFParser struct {
	ocr    ports.OCR
	always ports.OCR
}

// PDFOption customises a PDFParser.
type PDFOption func(*PDFParser)

// WithAlwaysOCR sends every page, including those with a usable text layer,
// to engine. When engine fails or returns nothing for such a page, its text
// layer is kept.
func WithAlwaysOCR(engine ports.OCR) PDFOption {
	return func(p *PDFParser) { p.always = engine }
}

// NewPDFParser builds a PDF parser. ocr may be nil: pages that need OCR are
// then skipped with a warning.
func NewPDFParser(ocr ports.OCR, opts ...PDFOption) *PDFParser {
	p := &PDFParser{ocr: ocr}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type pageResult struct {
	text   string
	source string
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

	results := make([]pageResult, pageCount)
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentOCR)
dispatch:
	for i, text := range cleaned {
		usable := !needsOCR(text)
		var engine ports.OCR
		switch {
		case forceOCR || !usable:
			engine = p.ocr
		case p.always != nil:
			engine = p.always
		default:
			results[i] = pageResult{text: text, source: "text"}
			continue
		}
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break dispatch
		}
		wg.Add(1)
		go func(i int, text string, usable bool, engine ports.OCR) {
			defer wg.Done()
			defer func() { <-sem }()
			pageNum := i + 1
			if ocrText := p.ocrPage(ctx, engine, inputPath, tmpDir, pageNum); ocrText != "" {
				results[i] = pageResult{text: ocrText, source: "ocr"}
				return
			}
			if usable && text != "" {
				slog.Warn("pdf: OCR gave no text; keeping the text layer", "page", pageNum)
				results[i] = pageResult{text: text, source: "text"}
			}
		}(i, text, usable, engine)
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("parse pdf: %w", err)
	}

	var docs []schema.Document
	for i, r := range results {
		if r.text == "" {
			continue
		}
		docs = append(docs, schema.Document{
			PageContent: r.text,
			Metadata: map[string]any{
				"page_number": i + 1,
				"page_count":  pageCount,
				"source":      r.source,
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
func (p *PDFParser) ocrPage(ctx context.Context, engine ports.OCR, pdfPath, tmpDir string, page int) string {
	if engine == nil {
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
	text, err := engine.Recognize(ctx, png)
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
