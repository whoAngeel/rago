package ports

import "io"

type BlobStorage interface {
	Save(filename string, reader io.Reader) (path string, err error)
	Get(path string) (io.ReadCloser, error)
	Delete(path string) error
}
