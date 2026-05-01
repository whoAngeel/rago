package ports

import (
	"context"
	"io"
)

type Parser interface {
	Parse(ctx context.Context, reader io.Reader, contentType string) (string, error)
}
