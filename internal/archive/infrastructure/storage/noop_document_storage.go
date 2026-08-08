package storage

import (
	"context"
	"fmt"
)

// NoopDocumentStorage is a stub DocumentStorage until iteration C.
type NoopDocumentStorage struct{}

// NewNoopDocumentStorage creates the stub.
func NewNoopDocumentStorage() *NoopDocumentStorage {
	return &NoopDocumentStorage{}
}

// Put is not supported in iteration A.
func (s *NoopDocumentStorage) Put(_ context.Context, _ string, _ string, _ []byte) error {
	return fmt.Errorf("document storage not enabled yet")
}

// Get is not supported in iteration A.
func (s *NoopDocumentStorage) Get(_ context.Context, _ string) ([]byte, string, error) {
	return nil, "", fmt.Errorf("document storage not enabled yet")
}

// Delete is not supported in iteration A.
func (s *NoopDocumentStorage) Delete(_ context.Context, _ string) error {
	return fmt.Errorf("document storage not enabled yet")
}
