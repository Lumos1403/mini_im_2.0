package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidStoragePath = errors.New("invalid storage path")

type LocalFileStorage struct {
	root string
}

func NewLocalFileStorage(root string) (*LocalFileStorage, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		root = "./uploads"
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return nil, err
	}

	return &LocalFileStorage{root: filepath.Clean(absRoot)}, nil
}

func (s *LocalFileStorage) Save(ctx context.Context, file multipart.File, storagePath string) (*StoredFile, error) {
	fullPath, cleanPath, err := s.resolve(storagePath)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	dst, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return nil, err
	}
	removePartial := true
	defer func() {
		_ = dst.Close()
		if removePartial {
			_ = os.Remove(fullPath)
		}
	}()

	hasher := sha256.New()
	size, err := copyWithHash(ctx, dst, file, hasher)
	if err != nil {
		return nil, err
	}

	removePartial = false
	return &StoredFile{
		StoragePath: filepath.ToSlash(cleanPath),
		Size:        size,
		SHA256:      hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

func (s *LocalFileStorage) Open(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fullPath, _, err := s.resolve(storagePath)
	if err != nil {
		return nil, err
	}
	return os.Open(fullPath)
}

func (s *LocalFileStorage) Delete(ctx context.Context, storagePath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath, _, err := s.resolve(storagePath)
	if err != nil {
		return err
	}
	err = os.Remove(fullPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *LocalFileStorage) resolve(storagePath string) (string, string, error) {
	storagePath = strings.TrimSpace(storagePath)
	if storagePath == "" || filepath.IsAbs(storagePath) {
		return "", "", ErrInvalidStoragePath
	}

	cleanPath := filepath.Clean(filepath.FromSlash(storagePath))
	if cleanPath == "." || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(os.PathSeparator)) {
		return "", "", ErrInvalidStoragePath
	}

	fullPath := filepath.Join(s.root, cleanPath)
	rel, err := filepath.Rel(s.root, fullPath)
	if err != nil {
		return "", "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", "", ErrInvalidStoragePath
	}

	return fullPath, cleanPath, nil
}

func copyWithHash(ctx context.Context, dst io.Writer, src io.Reader, hasher hash.Hash) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64
	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			chunk := buf[:nr]
			nw, ew := dst.Write(chunk)
			if nw > 0 {
				written += int64(nw)
				if _, err := hasher.Write(chunk[:nw]); err != nil {
					return written, err
				}
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				return written, nil
			}
			return written, er
		}
	}
}
