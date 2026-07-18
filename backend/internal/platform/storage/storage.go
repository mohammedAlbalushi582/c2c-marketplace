// Package storage abstracts blob storage behind an interface so the local
// filesystem implementation can later be swapped for S3 / Oracle Object Storage.
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage persists uploaded files and resolves their public URLs.
type Storage interface {
	// Save stores r under key and returns the public URL.
	Save(ctx context.Context, key string, r io.Reader) (string, error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

// LocalStorage writes files under baseDir and serves them at publicBaseURL/uploads/.
type LocalStorage struct {
	baseDir       string
	publicBaseURL string
}

func NewLocal(baseDir, publicBaseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	return &LocalStorage{baseDir: baseDir, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}, nil
}

func (l *LocalStorage) Save(_ context.Context, key string, r io.Reader) (string, error) {
	dest := filepath.Join(l.baseDir, filepath.Clean("/"+key))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return l.PublicURL(key), nil
}

func (l *LocalStorage) Delete(_ context.Context, key string) error {
	dest := filepath.Join(l.baseDir, filepath.Clean("/"+key))
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *LocalStorage) PublicURL(key string) string {
	return fmt.Sprintf("%s/uploads/%s", l.publicBaseURL, strings.TrimLeft(key, "/"))
}

// BaseDir exposes the on-disk root so the HTTP layer can serve it statically.
func (l *LocalStorage) BaseDir() string { return l.baseDir }
