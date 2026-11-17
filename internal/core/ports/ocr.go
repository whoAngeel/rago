package ports

import "context"

// OCR turns one rasterised page into text.
type OCR interface {
	Recognize(ctx context.Context, png []byte) (string, error)
}
