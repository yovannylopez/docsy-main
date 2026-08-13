package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

const (
	categoryLabelMaxLen      = 120
	categoryLabelMinLen      = 2
	maxCustomCategories      = 20
	customCategorySortBase   = 1000
	customCategoryCodePrefix = "c_"
	auditResourceCategory    = "document_category"
)

func toCategoryResponse(c entities.DocumentCategory) dtos.DocumentCategoryResponse {
	return dtos.DocumentCategoryResponse{
		Code:      c.Code,
		LabelES:   c.LabelES,
		SortOrder: c.SortOrder,
		IsSystem:  c.IsSystem,
	}
}

func normalizeCategoryLabel(raw string) (string, error) {
	label := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if label == "" {
		return "", domainerrors.ErrCategoryLabelRequired
	}
	if utf8.RuneCountInString(label) < categoryLabelMinLen {
		return "", domainerrors.ErrCategoryLabelRequired
	}
	if utf8.RuneCountInString(label) > categoryLabelMaxLen {
		return "", domainerrors.ErrCategoryLabelTooLong
	}
	return label, nil
}

func newCustomCategoryCode() string {
	return customCategoryCodePrefix + strings.ReplaceAll(uuid.NewString(), "-", "")
}

// ListCategoriesUseCase lists system + workspace custom categories.
type ListCategoriesUseCase struct {
	access documentAccess
}

// NewListCategoriesUseCase creates the use case.
func NewListCategoriesUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{access: documentAccess{workspaceRepo, ensureUC, docRepo}}
}

// Execute returns active categories visible in the workspace.
func (uc *ListCategoriesUseCase) Execute(ctx context.Context, userID, workspaceID string) ([]dtos.DocumentCategoryResponse, error) {
	ws, err := uc.access.workspaceForUser(ctx, userID, workspaceID, false)
	if err != nil {
		return nil, err
	}
	cats, err := uc.access.docRepo.ListCategories(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	out := make([]dtos.DocumentCategoryResponse, 0, len(cats))
	for _, c := range cats {
		out = append(out, toCategoryResponse(c))
	}
	return out, nil
}

// CreateCategoryUseCase creates a flat custom category.
type CreateCategoryUseCase struct {
	access    documentAccess
	auditRepo authports.AuditRepository
}

// NewCreateCategoryUseCase creates the use case.
func NewCreateCategoryUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	auditRepo authports.AuditRepository,
) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{
		access:    documentAccess{workspaceRepo, ensureUC, docRepo},
		auditRepo: auditRepo,
	}
}

// Execute creates a custom category for the workspace.
func (uc *CreateCategoryUseCase) Execute(
	ctx context.Context,
	userID string,
	req *dtos.CreateCategoryRequest,
) (*dtos.DocumentCategoryResponse, error) {
	if req == nil {
		return nil, domainerrors.ErrCategoryLabelRequired
	}
	ws, err := uc.access.workspaceForUser(ctx, userID, req.WorkspaceID, true)
	if err != nil {
		return nil, err
	}
	label, err := normalizeCategoryLabel(req.LabelES)
	if err != nil {
		return nil, err
	}
	n, err := uc.access.docRepo.CountCustomCategories(ctx, ws.ID)
	if err != nil {
		return nil, fmt.Errorf("count custom categories: %w", err)
	}
	if n >= maxCustomCategories {
		return nil, domainerrors.ErrTooManyCustomCategories
	}
	existing, err := uc.access.docRepo.ListCategories(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(label)
	for _, c := range existing {
		if strings.ToLower(c.LabelES) == lower {
			return nil, domainerrors.ErrCategoryDuplicateLabel
		}
	}

	now := time.Now().UTC()
	wsID := ws.ID
	cat := &entities.DocumentCategory{
		Code:        newCustomCategoryCode(),
		WorkspaceID: &wsID,
		LabelES:     label,
		SortOrder:   customCategorySortBase + n,
		IsActive:    true,
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := uc.access.docRepo.CreateCategory(ctx, cat); err != nil {
		if isUniqueViolation(err) {
			return nil, domainerrors.ErrCategoryDuplicateLabel
		}
		return nil, fmt.Errorf("create category: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveCategoryCreated,
		auditResourceCategory, cat.Code, "Archive custom category created successfully",
	)
	resp := toCategoryResponse(*cat)
	return &resp, nil
}

// UpdateCategoryUseCase renames a custom category.
type UpdateCategoryUseCase struct {
	access    documentAccess
	auditRepo authports.AuditRepository
}

// NewUpdateCategoryUseCase creates the use case.
func NewUpdateCategoryUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	auditRepo authports.AuditRepository,
) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{
		access:    documentAccess{workspaceRepo, ensureUC, docRepo},
		auditRepo: auditRepo,
	}
}

// Execute renames a custom category, or a system category when allowSystemEdit is true (super_admin).
func (uc *UpdateCategoryUseCase) Execute(
	ctx context.Context,
	userID, workspaceID, code string,
	req *dtos.UpdateCategoryRequest,
	allowSystemEdit bool,
) (*dtos.DocumentCategoryResponse, error) {
	if req == nil {
		return nil, domainerrors.ErrCategoryLabelRequired
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, domainerrors.ErrInvalidCategory
	}
	label, err := normalizeCategoryLabel(req.LabelES)
	if err != nil {
		return nil, err
	}

	wsID := strings.TrimSpace(workspaceID)
	if wsID == "" {
		wsID = strings.TrimSpace(req.WorkspaceID)
	}

	if allowSystemEdit {
		if sys := findSystemCategory(ctx, uc.access.docRepo, wsID, code); sys != nil {
			return uc.renameSystemCategory(ctx, userID, sys, label)
		}
	}

	ws, err := uc.access.workspaceForUser(ctx, userID, wsID, true)
	if err != nil {
		return nil, err
	}
	cat, err := uc.access.docRepo.FindCategory(ctx, ws.ID, code)
	if err != nil {
		return nil, domainerrors.ErrCategoryNotFound
	}
	if cat.IsSystem {
		return nil, domainerrors.ErrCannotModifySystemCategory
	}
	existing, err := uc.access.docRepo.ListCategories(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(label)
	for _, c := range existing {
		if c.Code == cat.Code {
			continue
		}
		if strings.ToLower(c.LabelES) == lower {
			return nil, domainerrors.ErrCategoryDuplicateLabel
		}
	}
	cat.LabelES = label
	cat.UpdatedAt = time.Now().UTC()
	if err := uc.access.docRepo.UpdateCategory(ctx, cat); err != nil {
		if isUniqueViolation(err) {
			return nil, domainerrors.ErrCategoryDuplicateLabel
		}
		return nil, fmt.Errorf("update category: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveCategoryUpdated,
		auditResourceCategory, cat.Code, "Archive custom category updated successfully",
	)
	resp := toCategoryResponse(*cat)
	return &resp, nil
}

func findSystemCategory(
	ctx context.Context,
	repo ports.DocumentRepository,
	workspaceID, code string,
) *entities.DocumentCategory {
	cats, err := repo.ListCategories(ctx, workspaceID)
	if err != nil {
		return nil
	}
	for i := range cats {
		if cats[i].Code == code && cats[i].IsSystem {
			c := cats[i]
			return &c
		}
	}
	return nil
}

func (uc *UpdateCategoryUseCase) renameSystemCategory(
	ctx context.Context,
	userID string,
	cat *entities.DocumentCategory,
	label string,
) (*dtos.DocumentCategoryResponse, error) {
	existing, err := uc.access.docRepo.ListCategories(ctx, "")
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(label)
	for _, c := range existing {
		if !c.IsSystem || c.Code == cat.Code {
			continue
		}
		if strings.ToLower(c.LabelES) == lower {
			return nil, domainerrors.ErrCategoryDuplicateLabel
		}
	}
	cat.LabelES = label
	cat.UpdatedAt = time.Now().UTC()
	if err := uc.access.docRepo.UpdateSystemCategory(ctx, cat); err != nil {
		return nil, fmt.Errorf("update system category: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveCategoryUpdated,
		auditResourceCategory, cat.Code, "Archive system category updated successfully",
	)
	resp := toCategoryResponse(*cat)
	return &resp, nil
}

// DeactivateCategoryUseCase soft-deactivates an unused custom category.
type DeactivateCategoryUseCase struct {
	access    documentAccess
	auditRepo authports.AuditRepository
}

// NewDeactivateCategoryUseCase creates the use case.
func NewDeactivateCategoryUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	auditRepo authports.AuditRepository,
) *DeactivateCategoryUseCase {
	return &DeactivateCategoryUseCase{
		access:    documentAccess{workspaceRepo, ensureUC, docRepo},
		auditRepo: auditRepo,
	}
}

// Execute deactivates a custom category when no documents reference it.
func (uc *DeactivateCategoryUseCase) Execute(ctx context.Context, userID, workspaceID, code string) error {
	ws, err := uc.access.workspaceForUser(ctx, userID, workspaceID, true)
	if err != nil {
		return err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return domainerrors.ErrInvalidCategory
	}
	cat, err := uc.access.docRepo.FindCategory(ctx, ws.ID, code)
	if err != nil {
		return domainerrors.ErrCategoryNotFound
	}
	if cat.IsSystem {
		return domainerrors.ErrCannotModifySystemCategory
	}
	counts, err := uc.access.docRepo.CountByCategory(ctx, ws.ID, entities.DocumentStatusAll)
	if err != nil {
		return fmt.Errorf("count by category: %w", err)
	}
	if counts[code] > 0 {
		return domainerrors.ErrCategoryInUse
	}
	if err := uc.access.docRepo.DeactivateCategory(ctx, ws.ID, code); err != nil {
		return fmt.Errorf("deactivate category: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveCategoryDeactivated,
		auditResourceCategory, code, "Archive custom category deactivated successfully",
	)
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
