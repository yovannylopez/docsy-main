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
	// CategoryExists reports whether code is an active system category or an active custom one for the workspace.
	CategoryExists(ctx context.Context, workspaceID, code string) (bool, error)
	// ListCategories returns active system categories plus custom ones for the workspace.
	ListCategories(ctx context.Context, workspaceID string) ([]entities.DocumentCategory, error)
	FindCategory(ctx context.Context, workspaceID, code string) (*entities.DocumentCategory, error)
	CreateCategory(ctx context.Context, cat *entities.DocumentCategory) error
	UpdateCategory(ctx context.Context, cat *entities.DocumentCategory) error
	// UpdateSystemCategory updates the label of an active system category (global seed).
	UpdateSystemCategory(ctx context.Context, cat *entities.DocumentCategory) error
	DeactivateCategory(ctx context.Context, workspaceID, code string) error
	CountCustomCategories(ctx context.Context, workspaceID string) (int, error)
	// CountByCategory returns document counts keyed by category_code for a workspace.
	CountByCategory(ctx context.Context, workspaceID string, status string) (map[string]int, error)
	// CountDueAlerts returns upcoming (today..+7d) and expired (before today) counts for documents with due_date.
	CountDueAlerts(ctx context.Context, workspaceID, status string) (upcoming, expired int, err error)
	// CountDueAlertsByCategory returns upcoming/expired due counts keyed by category_code.
	CountDueAlertsByCategory(ctx context.Context, workspaceID, status string) (map[string]dtos.CategoryDueAlertCounts, error)
}

// ListDocumentsService lists documents for the caller's personal workspace.
type ListDocumentsService interface {
	Execute(ctx context.Context, userID string, filter dtos.ListDocumentsFilter) ([]dtos.DocumentResponse, int, error)
	// CountDueAlerts returns in-app due summary counts for the caller's workspace.
	CountDueAlerts(ctx context.Context, userID, workspaceID, status string) (upcoming, expired int, err error)
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

// ListCategoriesService lists active document categories for a workspace (system + custom).
type ListCategoriesService interface {
	Execute(ctx context.Context, userID, workspaceID string) ([]dtos.DocumentCategoryResponse, error)
}

// CreateCategoryService creates a flat custom category in a workspace.
type CreateCategoryService interface {
	Execute(ctx context.Context, userID string, req *dtos.CreateCategoryRequest) (*dtos.DocumentCategoryResponse, error)
}

// UpdateCategoryService renames a category (custom, or system when allowSystemEdit).
type UpdateCategoryService interface {
	Execute(
		ctx context.Context,
		userID, workspaceID, code string,
		req *dtos.UpdateCategoryRequest,
		allowSystemEdit bool,
	) (*dtos.DocumentCategoryResponse, error)
}

// DeactivateCategoryService soft-deactivates a custom category when unused.
type DeactivateCategoryService interface {
	Execute(ctx context.Context, userID, workspaceID, code string) error
}

// ListCategoryFoldersService lists category folders with document counts for a workspace.
type ListCategoryFoldersService interface {
	Execute(ctx context.Context, userID string, workspaceID, status string) ([]dtos.CategoryFolderResponse, error)
}
