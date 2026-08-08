package ports

import "context"

// DocumentStorage stores binary attachments on a backend (local disk, object store, …).
type DocumentStorage interface {
	Put(ctx context.Context, key string, contentType string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, string, error)
	Delete(ctx context.Context, key string) error
}
