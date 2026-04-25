package storage

import (
	"context"
	"io"
	"mime/multipart"
)

type StoredFile struct {
	StoragePath string
	Size        int64
	SHA256      string
}

type FileStorage interface {
	Save(ctx context.Context, file multipart.File, storagePath string) (*StoredFile, error)
	Open(ctx context.Context, storagePath string) (io.ReadCloser, error)
	Delete(ctx context.Context, storagePath string) error
}
