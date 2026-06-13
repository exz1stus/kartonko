package storage

import (
	"context"
	"io"
	"time"
)

type FileInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

type Storage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Stat(ctx context.Context, key string) (*FileInfo, error)
	List(ctx context.Context, prefix string) ([]FileInfo, error)
}
