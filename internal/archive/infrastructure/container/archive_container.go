package container

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/archive/infrastructure/adapters"
	"github.com/yovannylopez/docsy-main/internal/archive/infrastructure/ocr"
	"github.com/yovannylopez/docsy-main/internal/archive/infrastructure/repositories"
	"github.com/yovannylopez/docsy-main/internal/archive/infrastructure/storage"
	"github.com/yovannylopez/docsy-main/internal/archive/transport/handlers"
	"github.com/yovannylopez/docsy-main/internal/archive/usecases"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	sharedconfig "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
)

// ArchiveContainer wires archive module dependencies.
type ArchiveContainer struct {
	WorkspaceRepo    ports.WorkspaceRepository
	DocumentRepo     ports.DocumentRepository
	DocumentStorage  ports.DocumentStorage
	EnsurePersonalUC ports.EnsurePersonalWorkspaceService
	StorageUsageUC   ports.GetStorageUsageService
	ListDocumentsUC  ports.ListDocumentsService
	Handler          *handlers.ArchiveHandler
	PageHandler      *handlers.ArchivePageHandler
}

// NewArchiveContainer creates and wires the archive module.
func NewArchiveContainer(
	db *sqlx.DB,
	storageCfg sharedconfig.StorageConfig,
	ocrCfg sharedconfig.OCRConfig,
	userFinder adapters.AuthUserFinder,
	auditRepo authports.AuditRepository,
) (*ArchiveContainer, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	wsRepo := repositories.NewWorkspaceRepositoryAdapter(db)
	docRepo := repositories.NewDocumentRepositoryAdapter(db)
	fileRepo := repositories.NewDocumentFileRepositoryAdapter(db)
	userDir := adapters.NewUserDirectoryAdapter(userFinder)

	docStorage, err := storage.NewLocalDocumentStorage(storageCfg.DocumentPath)
	if err != nil {
		return nil, fmt.Errorf("document storage: %w", err)
	}

	ocrExtractor := ocr.NewTesseractExtractor(ocrCfg)
	suggestOCRUC := usecases.NewSuggestDocumentFieldsUseCase(ocrExtractor, storageCfg.MaxFileSize, auditRepo)

	ensureUC := usecases.NewEnsurePersonalWorkspaceUseCase(wsRepo)

	listWorkspacesUC := usecases.NewListWorkspacesUseCase(wsRepo, ensureUC)
	createHouseholdUC := usecases.NewCreateHouseholdWorkspaceUseCase(wsRepo, auditRepo)
	listMembersUC := usecases.NewListWorkspaceMembersUseCase(wsRepo)
	inviteMemberUC := usecases.NewInviteWorkspaceMemberUseCase(wsRepo, userDir, auditRepo)
	updateMemberRoleUC := usecases.NewUpdateWorkspaceMemberRoleUseCase(wsRepo, auditRepo)
	removeMemberUC := usecases.NewRemoveWorkspaceMemberUseCase(wsRepo, auditRepo)

	listDocsUC := usecases.NewListDocumentsUseCase(wsRepo, ensureUC, docRepo, fileRepo)
	listFoldersUC := usecases.NewListCategoryFoldersUseCase(wsRepo, ensureUC, docRepo)
	getDocUC := usecases.NewGetDocumentUseCase(wsRepo, ensureUC, docRepo)
	createDocUC := usecases.NewCreateDocumentUseCase(wsRepo, ensureUC, docRepo, auditRepo)
	updateDocUC := usecases.NewUpdateDocumentUseCase(wsRepo, ensureUC, docRepo, auditRepo)
	archiveDocUC := usecases.NewArchiveDocumentUseCase(wsRepo, ensureUC, docRepo, auditRepo)
	listCatsUC := usecases.NewListCategoriesUseCase(wsRepo, ensureUC, docRepo)
	createCatUC := usecases.NewCreateCategoryUseCase(wsRepo, ensureUC, docRepo, auditRepo)
	updateCatUC := usecases.NewUpdateCategoryUseCase(wsRepo, ensureUC, docRepo, auditRepo)
	deactivateCatUC := usecases.NewDeactivateCategoryUseCase(wsRepo, ensureUC, docRepo, auditRepo)

	uploadFileUC := usecases.NewUploadDocumentFileUseCase(
		wsRepo, ensureUC, docRepo, fileRepo, docStorage, storageCfg.MaxFileSize, auditRepo,
	)
	createWithFileUC := usecases.NewCreateDocumentWithFileUseCase(createDocUC, uploadFileUC, docRepo)
	listFilesUC := usecases.NewListDocumentFilesUseCase(wsRepo, ensureUC, docRepo, fileRepo)
	downloadFileUC := usecases.NewDownloadDocumentFileUseCase(wsRepo, ensureUC, docRepo, fileRepo, docStorage)
	deleteFileUC := usecases.NewDeleteDocumentFileUseCase(wsRepo, ensureUC, docRepo, fileRepo, docStorage, auditRepo)
	storageUsageUC := usecases.NewGetStorageUsageUseCase(fileRepo, storageCfg.QuotaBytes)

	return &ArchiveContainer{
		WorkspaceRepo:    wsRepo,
		DocumentRepo:     docRepo,
		DocumentStorage:  docStorage,
		EnsurePersonalUC: ensureUC,
		StorageUsageUC:   storageUsageUC,
		ListDocumentsUC:  listDocsUC,
		Handler: handlers.NewArchiveHandler(
			ensureUC, listWorkspacesUC, createHouseholdUC, listMembersUC, inviteMemberUC,
			updateMemberRoleUC, removeMemberUC,
			listDocsUC, getDocUC, createDocUC, updateDocUC, archiveDocUC, listCatsUC,
			createCatUC, updateCatUC, deactivateCatUC,
			uploadFileUC, listFilesUC, downloadFileUC, deleteFileUC,
		),
		PageHandler: handlers.NewArchivePageHandler(
			ensureUC, listWorkspacesUC, createHouseholdUC, listMembersUC, inviteMemberUC, removeMemberUC,
			listDocsUC, listFoldersUC, getDocUC, createDocUC, createWithFileUC, updateDocUC, listCatsUC,
			createCatUC, updateCatUC, deactivateCatUC,
			uploadFileUC, listFilesUC, downloadFileUC, deleteFileUC, suggestOCRUC,
		),
	}, nil
}

// GetArchiveHandler returns the JSON API handler.
func (c *ArchiveContainer) GetArchiveHandler() *handlers.ArchiveHandler {
	return c.Handler
}

// GetArchivePageHandler returns the HTMX page handler.
func (c *ArchiveContainer) GetArchivePageHandler() *handlers.ArchivePageHandler {
	return c.PageHandler
}

// GetStorageUsageUC returns the sidebar storage usage use case.
func (c *ArchiveContainer) GetStorageUsageUC() ports.GetStorageUsageService {
	return c.StorageUsageUC
}

// GetListDocumentsUC returns the list-documents use case (also used for due-alert badges).
func (c *ArchiveContainer) GetListDocumentsUC() ports.ListDocumentsService {
	return c.ListDocumentsUC
}
