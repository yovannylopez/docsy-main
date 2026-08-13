package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// documentAccess resolves the caller's workspace for document operations.
type documentAccess struct {
	workspaceRepo ports.WorkspaceRepository
	ensureUC      ports.EnsurePersonalWorkspaceService
	docRepo       ports.DocumentRepository
}

func (a *documentAccess) workspaceForUser(
	ctx context.Context,
	userID, workspaceID string,
	needWrite bool,
) (*dtos.WorkspaceResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		ws, err := a.ensureUC.Execute(ctx, userID)
		if err != nil {
			return nil, err
		}
		if needWrite && ws.MemberRole == entities.WorkspaceRoleViewer {
			return nil, domainerrors.ErrInsufficientWorkspaceRole
		}
		return ws, nil
	}

	ws, err := a.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, domainerrors.ErrWorkspaceNotFound
	}
	member, err := a.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, domainerrors.ErrNotWorkspaceMember
	}
	if needWrite && member.Role == entities.WorkspaceRoleViewer {
		return nil, domainerrors.ErrInsufficientWorkspaceRole
	}
	return toWorkspaceResponse(ws, member.Role), nil
}

const (
	maxExtraFields   = 20
	maxExtraKeyLen   = 64
	maxExtraLabelLen = 80
	maxExtraValueLen = 255
)

func toDocumentResponse(doc *entities.Document, label string) *dtos.DocumentResponse {
	return &dtos.DocumentResponse{
		ID:              doc.ID,
		WorkspaceID:     doc.WorkspaceID,
		CategoryCode:    doc.CategoryCode,
		CategoryLabel:   label,
		Title:           doc.Title,
		DocumentDate:    doc.DocumentDate,
		DueDate:         doc.DueDate,
		Issuer:          doc.Issuer,
		ReferenceNumber: doc.ReferenceNumber,
		AmountCents:     doc.AmountCents,
		Currency:        doc.Currency,
		Notes:           doc.Notes,
		ExtraFields:     extraFieldsToDTO(doc.ExtraFields),
		Status:          doc.Status,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       doc.UpdatedAt,
	}
}

func extraFieldsToDTO(in entities.ExtraFields) []dtos.ExtraFieldDTO {
	if len(in) == 0 {
		return nil
	}
	out := make([]dtos.ExtraFieldDTO, 0, len(in))
	for _, f := range in {
		out = append(out, dtos.ExtraFieldDTO{Key: f.Key, Label: f.Label, Value: f.Value})
	}
	return out
}

func normalizeExtraFields(in []dtos.ExtraFieldDTO) (entities.ExtraFields, error) {
	if len(in) == 0 {
		return nil, nil
	}
	if len(in) > maxExtraFields {
		return nil, domainerrors.ErrTooManyExtraFields
	}
	out := make(entities.ExtraFields, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, f := range in {
		key := strings.TrimSpace(f.Key)
		label := strings.TrimSpace(f.Label)
		value := strings.TrimSpace(f.Value)
		if key == "" || label == "" || value == "" {
			continue
		}
		if len(key) > maxExtraKeyLen || len(label) > maxExtraLabelLen || len(value) > maxExtraValueLen {
			return nil, domainerrors.ErrInvalidExtraField
		}
		for _, r := range key {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
				continue
			}
			return nil, domainerrors.ErrInvalidExtraField
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, entities.ExtraField{Key: key, Label: label, Value: value})
		if len(out) > maxExtraFields {
			return nil, domainerrors.ErrTooManyExtraFields
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func (a *documentAccess) categoryLabels(ctx context.Context, workspaceID string) (map[string]string, error) {
	cats, err := a.docRepo.ListCategories(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(cats))
	for _, c := range cats {
		out[c.Code] = c.LabelES
	}
	return out, nil
}

func optionalTrimmed(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

// CreateDocumentUseCase creates a document in the personal workspace.
type CreateDocumentUseCase struct {
	access    documentAccess
	auditRepo authports.AuditRepository
}

// NewCreateDocumentUseCase creates the use case.
func NewCreateDocumentUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	auditRepo authports.AuditRepository,
) *CreateDocumentUseCase {
	return &CreateDocumentUseCase{
		access:    documentAccess{workspaceRepo, ensureUC, docRepo},
		auditRepo: auditRepo,
	}
}

// Execute validates and creates a document.
func (uc *CreateDocumentUseCase) Execute(ctx context.Context, userID string, req *dtos.CreateDocumentRequest) (*dtos.DocumentResponse, error) {
	if req == nil {
		return nil, domainerrors.ErrTitleRequired
	}
	ws, err := uc.access.workspaceForUser(ctx, userID, req.WorkspaceID, true)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, domainerrors.ErrTitleRequired
	}
	category := strings.TrimSpace(req.CategoryCode)
	if category == "" {
		return nil, domainerrors.ErrCategoryRequired
	}
	ok, err := uc.access.docRepo.CategoryExists(ctx, ws.ID, category)
	if err != nil {
		return nil, fmt.Errorf("validate category: %w", err)
	}
	if !ok {
		return nil, domainerrors.ErrInvalidCategory
	}

	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = entities.DefaultDocumentCurrency
	}
	extras, err := normalizeExtraFields(req.ExtraFields)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	createdBy := userID
	doc := &entities.Document{
		ID:              uuid.NewString(),
		WorkspaceID:     ws.ID,
		CategoryCode:    category,
		Title:           title,
		DocumentDate:    req.DocumentDate,
		DueDate:         req.DueDate,
		Issuer:          optionalTrimmed(req.Issuer),
		ReferenceNumber: optionalTrimmed(req.ReferenceNumber),
		AmountCents:     req.AmountCents,
		Currency:        currency,
		Notes:           optionalTrimmed(req.Notes),
		ExtraFields:     extras,
		Status:          entities.DocumentStatusActive,
		CreatedBy:       &createdBy,
		UpdatedBy:       &createdBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := uc.access.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveDocumentCreated,
		auditResourceDocument, doc.ID, "Archive document created successfully",
	)
	labels, _ := uc.access.categoryLabels(ctx, ws.ID)
	return toDocumentResponse(doc, labels[doc.CategoryCode]), nil
}

// ListDocumentsUseCase lists documents for the personal workspace.
type ListDocumentsUseCase struct {
	access   documentAccess
	fileRepo ports.DocumentFileRepository
}

// NewListDocumentsUseCase creates the use case.
func NewListDocumentsUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	fileRepo ports.DocumentFileRepository,
) *ListDocumentsUseCase {
	return &ListDocumentsUseCase{
		access:   documentAccess{workspaceRepo, ensureUC, docRepo},
		fileRepo: fileRepo,
	}
}

// Execute lists documents with filters.
func (uc *ListDocumentsUseCase) Execute(ctx context.Context, userID string, filter dtos.ListDocumentsFilter) ([]dtos.DocumentResponse, int, error) {
	ws, err := uc.access.workspaceForUser(ctx, userID, filter.WorkspaceID, false)
	if err != nil {
		return nil, 0, err
	}
	docs, total, err := uc.access.docRepo.List(ctx, ws.ID, filter)
	if err != nil {
		return nil, 0, err
	}
	labels, err := uc.access.categoryLabels(ctx, ws.ID)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]string, 0, len(docs))
	for i := range docs {
		ids = append(ids, docs[i].ID)
	}
	primaries := map[string]ports.DocumentPrimaryFile{}
	if uc.fileRepo != nil && len(ids) > 0 {
		primaries, err = uc.fileRepo.FindPrimaryByDocumentIDs(ctx, ids)
		if err != nil {
			return nil, 0, fmt.Errorf("load primary files: %w", err)
		}
	}

	out := make([]dtos.DocumentResponse, 0, len(docs))
	for i := range docs {
		resp := *toDocumentResponse(&docs[i], labels[docs[i].CategoryCode])
		if pf, ok := primaries[docs[i].ID]; ok {
			resp.PrimaryOriginalName = pf.OriginalName
			resp.PrimaryContentType = pf.ContentType
		}
		out = append(out, resp)
	}
	return out, total, nil
}

// GetDocumentUseCase retrieves one document.
type GetDocumentUseCase struct {
	access documentAccess
}

// NewGetDocumentUseCase creates the use case.
func NewGetDocumentUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
) *GetDocumentUseCase {
	return &GetDocumentUseCase{access: documentAccess{workspaceRepo, ensureUC, docRepo}}
}

// Execute returns a document by id.
func (uc *GetDocumentUseCase) Execute(ctx context.Context, userID, workspaceID, documentID string) (*dtos.DocumentResponse, error) {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return nil, domainerrors.ErrDocumentIDRequired
	}
	ws, err := uc.access.workspaceForUser(ctx, userID, workspaceID, false)
	if err != nil {
		return nil, err
	}
	doc, err := uc.access.docRepo.FindByID(ctx, ws.ID, documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domainerrors.ErrDocumentNotFound
	}
	labels, _ := uc.access.categoryLabels(ctx, ws.ID)
	return toDocumentResponse(doc, labels[doc.CategoryCode]), nil
}

// UpdateDocumentUseCase updates document metadata.
type UpdateDocumentUseCase struct {
	access    documentAccess
	auditRepo authports.AuditRepository
}

// NewUpdateDocumentUseCase creates the use case.
func NewUpdateDocumentUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	auditRepo authports.AuditRepository,
) *UpdateDocumentUseCase {
	return &UpdateDocumentUseCase{
		access:    documentAccess{workspaceRepo, ensureUC, docRepo},
		auditRepo: auditRepo,
	}
}

// Execute applies partial updates.
func (uc *UpdateDocumentUseCase) Execute(
	ctx context.Context,
	userID, workspaceID, documentID string,
	req *dtos.UpdateDocumentRequest,
) (*dtos.DocumentResponse, error) {
	if req == nil {
		return nil, domainerrors.ErrTitleRequired
	}
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return nil, domainerrors.ErrDocumentIDRequired
	}
	if workspaceID == "" {
		workspaceID = req.WorkspaceID
	}
	ws, err := uc.access.workspaceForUser(ctx, userID, workspaceID, true)
	if err != nil {
		return nil, err
	}
	doc, err := uc.access.docRepo.FindByID(ctx, ws.ID, documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domainerrors.ErrDocumentNotFound
	}
	if err := applyDocumentUpdate(ctx, &uc.access, doc, req); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	doc.UpdatedAt = now
	updatedBy := userID
	doc.UpdatedBy = &updatedBy

	if err := uc.access.docRepo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("update document: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveDocumentUpdated,
		auditResourceDocument, doc.ID, "Archive document updated successfully",
	)
	labels, _ := uc.access.categoryLabels(ctx, ws.ID)
	return toDocumentResponse(doc, labels[doc.CategoryCode]), nil
}

func applyDocumentUpdate(
	ctx context.Context,
	access *documentAccess,
	doc *entities.Document,
	req *dtos.UpdateDocumentRequest,
) error {
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return domainerrors.ErrTitleRequired
		}
		doc.Title = title
	}
	if req.CategoryCode != nil {
		cat := strings.TrimSpace(*req.CategoryCode)
		if cat == "" {
			return domainerrors.ErrCategoryRequired
		}
		ok, cErr := access.docRepo.CategoryExists(ctx, doc.WorkspaceID, cat)
		if cErr != nil {
			return cErr
		}
		if !ok {
			return domainerrors.ErrInvalidCategory
		}
		doc.CategoryCode = cat
	}
	if req.DocumentDate != nil {
		doc.DocumentDate = req.DocumentDate
	}
	if req.ClearDueDate {
		doc.DueDate = nil
	} else if req.DueDate != nil {
		doc.DueDate = req.DueDate
	}
	if req.Issuer != nil {
		doc.Issuer = optionalTrimmed(req.Issuer)
	}
	if req.ReferenceNumber != nil {
		doc.ReferenceNumber = optionalTrimmed(req.ReferenceNumber)
	}
	if req.ClearAmount {
		doc.AmountCents = nil
	} else if req.AmountCents != nil {
		doc.AmountCents = req.AmountCents
	}
	if req.Currency != nil {
		cur := strings.TrimSpace(*req.Currency)
		if cur != "" {
			doc.Currency = cur
		}
	}
	if req.Notes != nil {
		doc.Notes = optionalTrimmed(req.Notes)
	}
	if req.SetExtraFields {
		extras, nErr := normalizeExtraFields(req.ExtraFields)
		if nErr != nil {
			return nErr
		}
		doc.ExtraFields = extras
	}
	return nil
}

// ArchiveDocumentUseCase soft-archives a document.
type ArchiveDocumentUseCase struct {
	access    documentAccess
	auditRepo authports.AuditRepository
}

// NewArchiveDocumentUseCase creates the use case.
func NewArchiveDocumentUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	auditRepo authports.AuditRepository,
) *ArchiveDocumentUseCase {
	return &ArchiveDocumentUseCase{
		access:    documentAccess{workspaceRepo, ensureUC, docRepo},
		auditRepo: auditRepo,
	}
}

// Execute sets status to archived.
func (uc *ArchiveDocumentUseCase) Execute(ctx context.Context, userID, workspaceID, documentID string) (*dtos.DocumentResponse, error) {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return nil, domainerrors.ErrDocumentIDRequired
	}
	ws, err := uc.access.workspaceForUser(ctx, userID, workspaceID, true)
	if err != nil {
		return nil, err
	}
	doc, err := uc.access.docRepo.FindByID(ctx, ws.ID, documentID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domainerrors.ErrDocumentNotFound
	}
	doc.Status = entities.DocumentStatusArchived
	now := time.Now().UTC()
	doc.UpdatedAt = now
	updatedBy := userID
	doc.UpdatedBy = &updatedBy
	if err := uc.access.docRepo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("archive document: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveDocumentArchived,
		auditResourceDocument, doc.ID, "Archive document archived successfully",
	)
	labels, _ := uc.access.categoryLabels(ctx, ws.ID)
	return toDocumentResponse(doc, labels[doc.CategoryCode]), nil
}

// ListCategoryFoldersUseCase lists virtual category folders with document counts.
type ListCategoryFoldersUseCase struct {
	access documentAccess
}

// NewListCategoryFoldersUseCase creates the use case.
func NewListCategoryFoldersUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
) *ListCategoryFoldersUseCase {
	return &ListCategoryFoldersUseCase{access: documentAccess{workspaceRepo, ensureUC, docRepo}}
}

// Execute returns categories with counts for the caller's workspace.
func (uc *ListCategoryFoldersUseCase) Execute(
	ctx context.Context,
	userID string,
	workspaceID, status string,
) ([]dtos.CategoryFolderResponse, error) {
	ws, err := uc.access.workspaceForUser(ctx, userID, workspaceID, false)
	if err != nil {
		return nil, err
	}
	cats, err := uc.access.docRepo.ListCategories(ctx, ws.ID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	counts, err := uc.access.docRepo.CountByCategory(ctx, ws.ID, status)
	if err != nil {
		return nil, fmt.Errorf("count by category: %w", err)
	}
	out := make([]dtos.CategoryFolderResponse, 0, len(cats))
	for _, c := range cats {
		out = append(out, dtos.CategoryFolderResponse{
			Code:      c.Code,
			LabelES:   c.LabelES,
			SortOrder: c.SortOrder,
			Count:     counts[c.Code],
			IsSystem:  c.IsSystem,
		})
	}
	return out, nil
}
