package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/whoAngeel/rago/internal/core/ports"
)

// TesseractOCR implements ports.OCR with the local tesseract binary.
type TesseractOCR struct {
	languages []string
}

var _ ports.OCR = (*TesseractOCR)(nil)

// NewTesseractOCR builds a Tesseract-backed OCR; languages default to
// Spanish and English.
func NewTesseractOCR(langs ...string) *TesseractOCR {
	if len(langs) == 0 {
		langs = []string{"spa", "eng"}
	}
	return &TesseractOCR{languages: langs}
}

// Recognize runs tesseract over one PNG image and returns the raw text.
func (t *TesseractOCR) Recognize(ctx context.Context, png []byte) (string, error) {
	dir, err := os.MkdirTemp("", "tesseract-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "page.png")
	if err := os.WriteFile(path, png, 0o600); err != nil {
		return "", fmt.Errorf("write temp image: %w", err)
	}
	text, err := runTesseract(ctx, path, t.languages)
	if err != nil {
		return "", fmt.Errorf("tesseract: %w", err)
	}
	return strings.TrimSpace(text), nil
}
