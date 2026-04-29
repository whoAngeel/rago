package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type BlobStorageAdapter struct {
	UploadDir string
}

func NewBlobStorageAdapter(uploadDir string) *BlobStorageAdapter {
	return &BlobStorageAdapter{
		UploadDir: uploadDir,
	}
}

func (b *BlobStorageAdapter) Save(ctx context.Context, filename string, reader io.Reader) (string, error) {
	if err := os.MkdirAll(b.UploadDir, os.ModePerm); err != nil {
		return "", err
	}
	uniqueName := fmt.Sprintf("%d_%s", time.Now().Unix(), filename)

	file, err := os.Create(filepath.Join(b.UploadDir, uniqueName))
	if err != nil {
		return "", err
	}
	file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return "", err
	}

	return uniqueName, nil
}

func (b *BlobStorageAdapter) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(b.UploadDir, path))
}

func (b *BlobStorageAdapter) Delete(ctx context.Context, path string) error {
	return os.Remove(filepath.Join(b.UploadDir, path))
}
