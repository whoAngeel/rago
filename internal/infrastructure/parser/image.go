package parser

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

type ImageParser struct {
	languages []string
}

func NewImageParser(langs ...string) *ImageParser {
	if len(langs) == 0 {
		langs = []string{"spa", "eng"}
	}
	return &ImageParser{languages: langs}
}

func (p *ImageParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	tmpDir, err := os.MkdirTemp("", "ocr-image-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ext := extensionFromContentType(contentType)
	inputPath := filepath.Join(tmpDir, "input"+ext)
	f, err := os.Create(inputPath)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	if _, err := io.Copy(f, reader); err != nil {
		f.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	f.Close()

	text, err := runTesseract(ctx, inputPath, p.languages)
	if err != nil {
		return nil, fmt.Errorf("ocr: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("no text extracted from image")
	}

	return []schema.Document{{
		PageContent: text,
		Metadata:    map[string]any{"content_type": contentType, "source": "ocr"},
	}}, nil
}

func runTesseract(ctx context.Context, imagePath string, langs []string) (string, error) {
	outBase := imagePath + "-out"
	args := []string{
		imagePath,
		outBase,
		"-l", strings.Join(langs, "+"),
	}
	cmd := exec.CommandContext(ctx, "tesseract", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w: %s", err, string(out))
	}
	data, err := os.ReadFile(outBase + ".txt")
	if err != nil {
		return "", fmt.Errorf("read tesseract output: %w", err)
	}
	return string(data), nil
}

func extensionFromContentType(ct string) string {
	switch ct {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/tiff":
		return ".tiff"
	default:
		return ".img"
	}
}
