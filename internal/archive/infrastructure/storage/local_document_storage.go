package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	storageDirPerm  = 0o750
	storageFilePerm = 0o600
)

// LocalDocumentStorage stores binaries under a local directory.
type LocalDocumentStorage struct {
	basePath string
}

// NewLocalDocumentStorage creates a disk-backed DocumentStorage.
func NewLocalDocumentStorage(basePath string) (*LocalDocumentStorage, error) {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		return nil, fmt.Errorf("document storage path is required")
	}
	if err := os.MkdirAll(basePath, storageDirPerm); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	abs, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}
	return &LocalDocumentStorage{basePath: abs}, nil
}

// Put writes data for the given relative key.
func (s *LocalDocumentStorage) Put(_ context.Context, key string, _ string, data []byte) error {
	full, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), storageDirPerm); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	if err := os.WriteFile(full, data, storageFilePerm); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// Get reads data for the given relative key. The returned content type is empty (metadata lives in DB).
func (s *LocalDocumentStorage) Get(_ context.Context, key string) ([]byte, string, error) {
	full, err := s.resolve(key)
	if err != nil {
		return nil, "", err
	}
	// full is constrained under basePath by resolve().
	data, err := os.ReadFile(full) //nolint:gosec // path validated against basePath
	if err != nil {
		return nil, "", fmt.Errorf("read file: %w", err)
	}
	return data, "", nil
}

// Delete removes the file for the given relative key. Missing files are ignored.
func (s *LocalDocumentStorage) Delete(_ context.Context, key string) error {
	full, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (s *LocalDocumentStorage) resolve(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("storage key is required")
	}
	clean := filepath.Clean(key)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid storage key")
	}
	full := filepath.Join(s.basePath, clean)
	rel, err := filepath.Rel(s.basePath, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid storage key")
	}
	return full, nil
}
