package entities

import "time"

// DocumentFile is a binary attachment for a document.
type DocumentFile struct {
	ID           string    `json:"id" db:"id"`
	DocumentID   string    `json:"document_id" db:"document_id"`
	StorageKey   string    `json:"storage_key" db:"storage_key"`
	OriginalName string    `json:"original_name" db:"original_name"`
	ContentType  string    `json:"content_type" db:"content_type"`
	SizeBytes    int64     `json:"size_bytes" db:"size_bytes"`
	UploadedBy   *string   `json:"uploaded_by,omitempty" db:"uploaded_by"`
	UploadedAt   time.Time `json:"uploaded_at" db:"uploaded_at"`
}
