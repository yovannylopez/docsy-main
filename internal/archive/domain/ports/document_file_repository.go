package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

// DocumentPrimaryFile is the newest attachment summary for a document card.
type DocumentPrimaryFile struct {
	DocumentID   string
	OriginalName string
	ContentType  string
}

// DocumentFileRepository persists document attachment metadata.
type DocumentFileRepository interface {
	Create(ctx context.Context, file *entities.DocumentFile) error
	ListByDocument(ctx context.Context, documentID string) ([]entities.DocumentFile, error)
	FindByID(ctx context.Context, documentID, fileID string) (*entities.DocumentFile, error)
	Delete(ctx context.Context, documentID, fileID string) error
	CountByDocument(ctx context.Context, documentID string) (int, error)
	// SumSizeBytesForUser returns total attachment bytes across workspaces the user belongs to.
	SumSizeBytesForUser(ctx context.Context, userID string) (int64, error)
	// FindPrimaryByDocumentIDs returns the newest file per document id.
	FindPrimaryByDocumentIDs(ctx context.Context, documentIDs []string) (map[string]DocumentPrimaryFile, error)
}

// GetStorageUsageService returns used bytes vs configured soft quota for the sidebar.
type GetStorageUsageService interface {
	Execute(ctx context.Context, userID string) (*dtos.StorageUsageResponse, error)
}

// UploadDocumentFileService uploads a binary attachment to a document.
type UploadDocumentFileService interface {
	Execute(ctx context.Context, userID, workspaceID, documentID, originalName, contentType string, data []byte) (*dtos.DocumentFileResponse, error)
}

// ListDocumentFilesService lists attachments for a document.
type ListDocumentFilesService interface {
	Execute(ctx context.Context, userID, workspaceID, documentID string) ([]dtos.DocumentFileResponse, error)
}

// DownloadDocumentFileService loads an attachment payload for download.
type DownloadDocumentFileService interface {
	Execute(ctx context.Context, userID, workspaceID, documentID, fileID string) (*dtos.DownloadDocumentFileResult, error)
}

// DeleteDocumentFileService removes an attachment from storage and metadata.
type DeleteDocumentFileService interface {
	Execute(ctx context.Context, userID, workspaceID, documentID, fileID string) error
}
