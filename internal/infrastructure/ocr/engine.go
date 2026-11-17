// Package ocr selects the OCR engine from configuration.
package ocr

import (
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/openrouter"
	"github.com/whoAngeel/rago/internal/infrastructure/parser"
)

// NewEngine returns the OCR engine chosen by cfg.OCREngine: "llm" uses the
// OpenRouter model with Tesseract as fallback (plain Tesseract when there is
// no API key); anything else uses Tesseract.
func NewEngine(cfg *config.Config) ports.OCR {
	if cfg.OCREngine == "llm" && cfg.OpenRouterKey != "" {
		return parser.NewFallbackOCR(newLLM(cfg), parser.NewTesseractOCR())
	}
	return parser.NewTesseractOCR()
}

// PDFOptions returns the parser options for the PDF OCR mode: the LLM engine
// alone for every page, only when the mode is "always", the engine is "llm"
// and an API key is set; otherwise nil (pages without a text layer only).
func PDFOptions(cfg *config.Config) []parser.PDFOption {
	if cfg.PDFOCRMode == "always" && cfg.OCREngine == "llm" && cfg.OpenRouterKey != "" {
		return []parser.PDFOption{parser.WithAlwaysOCR(newLLM(cfg))}
	}
	return nil
}

func newLLM(cfg *config.Config) ports.OCR {
	return openrouter.NewOCRAdapter(cfg.OpenRouterKey, cfg.OpenRouterBaseUrl, cfg.OCRModel)
}

// Effective reports the engine, model (empty for Tesseract) and PDF OCR mode
// that NewEngine and PDFOptions actually apply for cfg.
func Effective(cfg *config.Config) (engine, model, mode string) {
	engine, mode = "tesseract", "auto"
	if cfg.OCREngine == "llm" && cfg.OpenRouterKey != "" {
		engine, model = "llm", cfg.OCRModel
	}
	if len(PDFOptions(cfg)) > 0 {
		mode = "always"
	}
	return engine, model, mode
}
