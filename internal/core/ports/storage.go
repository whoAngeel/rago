package ports

import (
	"context"
	"io"
)

type BlobStorage interface {
	Upload(ctx context.Context, filename string, reader io.Reader) (path string, err error)
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
}
