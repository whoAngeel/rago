package parser

import (
	"context"
	"log/slog"

	"github.com/whoAngeel/rago/internal/core/ports"
)

type fallbackOCR struct {
	primary, secondary ports.OCR
}

// NewFallbackOCR returns an OCR that uses secondary when primary fails.
// An empty but successful primary result is returned as is.
func NewFallbackOCR(primary, secondary ports.OCR) ports.OCR {
	return &fallbackOCR{primary: primary, secondary: secondary}
}

func (f *fallbackOCR) Recognize(ctx context.Context, png []byte) (string, error) {
	text, err := f.primary.Recognize(ctx, png)
	if err == nil {
		return text, nil
	}
	slog.Warn("ocr: primary engine failed; falling back", "err", err)
	return f.secondary.Recognize(ctx, png)
}
