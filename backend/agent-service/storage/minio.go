package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client *minio.Client
	bucket string
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func NewMinioClient(cfg *MinioConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinioClient{client: client, bucket: cfg.Bucket}, nil
}

func (m *MinioClient) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}
	return fmt.Sprintf("minio://%s/%s", m.bucket, objectName), nil
}

func (m *MinioClient) GetURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}
	return url.String(), nil
}

func (m *MinioClient) Delete(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
}
