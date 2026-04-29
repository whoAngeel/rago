package ports

import (
	"context"
	"io"
)

type BlobStorage interface {
	Save(ctx context.Context, filename string, reader io.Reader) (path string, err error)
	Get(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}
