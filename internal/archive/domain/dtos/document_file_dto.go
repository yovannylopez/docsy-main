package dtos

import "time"

// DocumentFileResponse is the API/view DTO for an attachment.
type DocumentFileResponse struct {
	ID           string    `json:"id"`
	DocumentID   string    `json:"document_id"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// DownloadDocumentFileResult holds binary payload for download.
type DownloadDocumentFileResult struct {
	File *DocumentFileResponse
	Data []byte
}

// StorageUsageResponse is used/quota for the sidebar storage indicator.
type StorageUsageResponse struct {
	UsedBytes  int64 `json:"used_bytes"`
	QuotaBytes int64 `json:"quota_bytes"`
	Percent    int   `json:"percent"`
}
