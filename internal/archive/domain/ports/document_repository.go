package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

// DocumentRepository persists archive documents and categories.
type DocumentRepository interface {
	List(ctx context.Context, workspaceID string, filter dtos.ListDocumentsFilter) ([]entities.Document, int, error)
	FindByID(ctx context.Context, workspaceID, documentID string) (*entities.Document, error)
	Create(ctx context.Context, doc *entities.Document) error
	Update(ctx context.Context, doc *entities.Document) error
	// Delete hard-deletes a document (attachments cascade). Used to compensate failed create+attach.
	Delete(ctx context.Context, workspaceID, documentID string) error
	CategoryExists(ctx context.Context, code string) (bool, error)
	ListCategories(ctx context.Context) ([]entities.DocumentCategory, error)
	// CountByCategory returns document counts keyed by category_code for a workspace.
	CountByCategory(ctx context.Context, workspaceID string, status string) (map[string]int, error)
}

// ListDocumentsService lists documents for the caller's personal workspace.
type ListDocumentsService interface {
	Execute(ctx context.Context, userID string, filter dtos.ListDocumentsFilter) ([]dtos.DocumentResponse, int, error)
}

// GetDocumentService retrieves one document by id in the caller's workspace.
type GetDocumentService interface {
	Execute(ctx context.Context, userID, workspaceID, documentID string) (*dtos.DocumentResponse, error)
}

// CreateDocumentService creates a document in the caller's workspace.
type CreateDocumentService interface {
	Execute(ctx context.Context, userID string, req *dtos.CreateDocumentRequest) (*dtos.DocumentResponse, error)
}

// CreateDocumentWithFileService creates a document and uploads its first required attachment.
type CreateDocumentWithFileService interface {
	Execute(
		ctx context.Context,
		userID string,
		req *dtos.CreateDocumentRequest,
		originalName, contentType string,
		data []byte,
	) (*dtos.DocumentResponse, *dtos.DocumentFileResponse, error)
}

// UpdateDocumentService updates a document in the caller's workspace.
type UpdateDocumentService interface {
	Execute(ctx context.Context, userID, workspaceID, documentID string, req *dtos.UpdateDocumentRequest) (*dtos.DocumentResponse, error)
}

// ArchiveDocumentService soft-archives a document.
type ArchiveDocumentService interface {
	Execute(ctx context.Context, userID, workspaceID, documentID string) (*dtos.DocumentResponse, error)
}

// ListCategoriesService lists active document categories.
type ListCategoriesService interface {
	Execute(ctx context.Context) ([]dtos.DocumentCategoryResponse, error)
}

// ListCategoryFoldersService lists category folders with document counts for a workspace.
type ListCategoryFoldersService interface {
	Execute(ctx context.Context, userID string, workspaceID, status string) ([]dtos.CategoryFolderResponse, error)
}
