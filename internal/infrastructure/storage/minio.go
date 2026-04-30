package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioAdapter struct {
	client *minio.Client
	bucket string
}

func NewMinioAdapter(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioAdapter, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)

	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &MinioAdapter{client: client, bucket: bucket}, nil
}

func (m *MinioAdapter) Upload(ctx context.Context, objectKey string, reader io.Reader) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	return objectKey, nil
}

func (m *MinioAdapter) Download(ctx context.Context, objKey string) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, m.bucket, objKey, minio.GetObjectOptions{})
}

func (m *MinioAdapter) Delete(ctx context.Context, objKey string) error {
	return m.client.RemoveObject(ctx, m.bucket, objKey, minio.RemoveObjectOptions{})
}

func (m *MinioAdapter) Exists(ctx context.Context, objKey string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucket, objKey, minio.GetObjectOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}
