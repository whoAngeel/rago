package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/tmc/langchaingo/schema"
)

type PDFParser struct {
	gotenbergURL string
	httpClient   *http.Client
}

func NewPDFParser(gotenbergURL string) *PDFParser {
	return &PDFParser{
		gotenbergURL: gotenbergURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
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

	nativeText, err := extractPDFText(inputPath, tmpDir)
	if err == nil && len(strings.TrimSpace(nativeText)) > 0 {
		return splitByPages(nativeText), nil
	}

	if p.gotenbergURL == "" {
		return nil, fmt.Errorf("PDF has no extractable text and no Gotenberg endpoint configured")
	}

	ocrOutputPath := filepath.Join(tmpDir, "ocr_output.pdf")
	if err := p.runGotenbergOCR(ctx, inputPath, ocrOutputPath); err != nil {
		return nil, fmt.Errorf("gotenberg OCR failed: %w", err)
	}

	ocrText, err := extractPDFText(ocrOutputPath, tmpDir)
	if err != nil {
		return nil, err
	}
	return splitByPages(ocrText), nil
}

func extractPDFText(inputPath, tmpDir string) (string, error) {
	outDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	if err := api.ExtractContentFile(inputPath, outDir, nil, nil); err != nil {
		return "", err
	}

	contentFile := filepath.Join(outDir, "content.txt")
	data, err := os.ReadFile(contentFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *PDFParser) runGotenbergOCR(ctx context.Context, inputPath, outputPath string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("files", filepath.Base(inputPath))
	if err != nil {
		return err
	}
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, f); err != nil {
		f.Close()
		return err
	}
	f.Close()
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", p.gotenbergURL+"/forms/ocr", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotenberg returned %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func splitByPages(text string) []schema.Document {
	pages := strings.Split(text, "\x0c")
	var docs []schema.Document
	for i, page := range pages {
		page = strings.TrimSpace(page)
		if page == "" {
			continue
		}
		docs = append(docs, schema.Document{
			PageContent: page,
			Metadata: map[string]any{
				"page_number": i + 1,
				"page_count":  len(pages),
			},
		})
	}
	return docs
}
