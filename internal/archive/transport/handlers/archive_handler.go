package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// ArchiveHandler serves JSON API endpoints for the archive module.
type ArchiveHandler struct {
	ensurePersonalUC   ports.EnsurePersonalWorkspaceService
	listWorkspacesUC   ports.ListWorkspacesService
	createHouseholdUC  ports.CreateHouseholdWorkspaceService
	listMembersUC      ports.ListWorkspaceMembersService
	inviteMemberUC     ports.InviteWorkspaceMemberService
	updateMemberRoleUC ports.UpdateWorkspaceMemberRoleService
	removeMemberUC     ports.RemoveWorkspaceMemberService
	listDocsUC         ports.ListDocumentsService
	getDocUC           ports.GetDocumentService
	createDocUC        ports.CreateDocumentService
	updateDocUC        ports.UpdateDocumentService
	archiveDocUC       ports.ArchiveDocumentService
	listCatsUC         ports.ListCategoriesService
	createCatUC        ports.CreateCategoryService
	updateCatUC        ports.UpdateCategoryService
	deactivateCatUC    ports.DeactivateCategoryService
	uploadFileUC       ports.UploadDocumentFileService
	listFilesUC        ports.ListDocumentFilesService
	downloadFileUC     ports.DownloadDocumentFileService
	deleteFileUC       ports.DeleteDocumentFileService
}

// NewArchiveHandler creates the JSON handler.
func NewArchiveHandler(
	ensurePersonalUC ports.EnsurePersonalWorkspaceService,
	listWorkspacesUC ports.ListWorkspacesService,
	createHouseholdUC ports.CreateHouseholdWorkspaceService,
	listMembersUC ports.ListWorkspaceMembersService,
	inviteMemberUC ports.InviteWorkspaceMemberService,
	updateMemberRoleUC ports.UpdateWorkspaceMemberRoleService,
	removeMemberUC ports.RemoveWorkspaceMemberService,
	listDocsUC ports.ListDocumentsService,
	getDocUC ports.GetDocumentService,
	createDocUC ports.CreateDocumentService,
	updateDocUC ports.UpdateDocumentService,
	archiveDocUC ports.ArchiveDocumentService,
	listCatsUC ports.ListCategoriesService,
	createCatUC ports.CreateCategoryService,
	updateCatUC ports.UpdateCategoryService,
	deactivateCatUC ports.DeactivateCategoryService,
	uploadFileUC ports.UploadDocumentFileService,
	listFilesUC ports.ListDocumentFilesService,
	downloadFileUC ports.DownloadDocumentFileService,
	deleteFileUC ports.DeleteDocumentFileService,
) *ArchiveHandler {
	return &ArchiveHandler{
		ensurePersonalUC:   ensurePersonalUC,
		listWorkspacesUC:   listWorkspacesUC,
		createHouseholdUC:  createHouseholdUC,
		listMembersUC:      listMembersUC,
		inviteMemberUC:     inviteMemberUC,
		updateMemberRoleUC: updateMemberRoleUC,
		removeMemberUC:     removeMemberUC,
		listDocsUC:         listDocsUC,
		getDocUC:           getDocUC,
		createDocUC:        createDocUC,
		updateDocUC:        updateDocUC,
		archiveDocUC:       archiveDocUC,
		listCatsUC:         listCatsUC,
		createCatUC:        createCatUC,
		updateCatUC:        updateCatUC,
		deactivateCatUC:    deactivateCatUC,
		uploadFileUC:       uploadFileUC,
		listFilesUC:        listFilesUC,
		downloadFileUC:     downloadFileUC,
		deleteFileUC:       deleteFileUC,
	}
}

// GetMyWorkspace handles GET /api/v1/archive/workspaces/me
func (h *ArchiveHandler) GetMyWorkspace(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}

	ws, err := h.ensurePersonalUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}

	return responses.OK(c, ws, "workspace retrieved successfully")
}

// ListWorkspaces handles GET /api/v1/archive/workspaces
func (h *ArchiveHandler) ListWorkspaces(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	workspaces, err := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, workspaces, "workspaces retrieved successfully")
}

// CreateHousehold handles POST /api/v1/archive/workspaces/household
func (h *ArchiveHandler) CreateHousehold(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	var req dtos.CreateHouseholdRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}
	ws, err := h.createHouseholdUC.Execute(c.Request().Context(), userID, &req)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.Created(c, ws, "household workspace created successfully")
}

// ListMembers handles GET /api/v1/archive/workspaces/:id/members
func (h *ArchiveHandler) ListMembers(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	members, err := h.listMembersUC.Execute(c.Request().Context(), userID, c.Param("id"))
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, members, "members retrieved successfully")
}

// InviteMember handles POST /api/v1/archive/workspaces/:id/members
func (h *ArchiveHandler) InviteMember(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	var req dtos.InviteMemberRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}
	member, err := h.inviteMemberUC.Execute(c.Request().Context(), userID, c.Param("id"), &req)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.Created(c, member, "member invited successfully")
}

// UpdateMemberRole handles PATCH /api/v1/archive/workspaces/:id/members/:userId
func (h *ArchiveHandler) UpdateMemberRole(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	var req dtos.UpdateMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}
	member, err := h.updateMemberRoleUC.Execute(
		c.Request().Context(), userID, c.Param("id"), c.Param("userId"), req.Role,
	)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, member, "member role updated successfully")
}

// RemoveMember handles DELETE /api/v1/archive/workspaces/:id/members/:userId
func (h *ArchiveHandler) RemoveMember(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	if err := h.removeMemberUC.Execute(c.Request().Context(), userID, c.Param("id"), c.Param("userId")); err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, nil, "member removed successfully")
}

// ListDocuments handles GET /api/v1/archive/documents
func (h *ArchiveHandler) ListDocuments(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}

	params, err := pagination.NewDefaultParser().ParseFromQuery(c.QueryParam("limit"), c.QueryParam("offset"))
	if err != nil {
		return responses.BadRequest(c, err.Error())
	}

	filter := dtos.ListDocumentsFilter{
		WorkspaceID: workspaceIDParam(c),
		Category:    c.QueryParam("category"),
		Query:       c.QueryParam("q"),
		Status:      c.QueryParam("status"),
		Limit:       params.Limit,
		Offset:      params.Offset,
	}
	if from, ok, pErr := parseOptionalDate(c.QueryParam("from")); pErr != nil {
		return responses.BadRequest(c, "invalid from date")
	} else if ok {
		filter.From = from
	}
	if to, ok, pErr := parseOptionalDate(c.QueryParam("to")); pErr != nil {
		return responses.BadRequest(c, "invalid to date")
	} else if ok {
		filter.To = to
	}
	if due, ok, pErr := parseOptionalDate(c.QueryParam("due_before")); pErr != nil {
		return responses.BadRequest(c, "invalid due_before date")
	} else if ok {
		filter.DueBefore = due
	}
	if dueFrom, ok, pErr := parseOptionalDate(c.QueryParam("due_from")); pErr != nil {
		return responses.BadRequest(c, "invalid due_from date")
	} else if ok {
		filter.DueFrom = dueFrom
	}
	if dueTo, ok, pErr := parseOptionalDate(c.QueryParam("due_to")); pErr != nil {
		return responses.BadRequest(c, "invalid due_to date")
	} else if ok {
		filter.DueTo = dueTo
	}
	if filter.From != nil && filter.To != nil && filter.To.Before(*filter.From) {
		return responses.BadRequest(c, "to date must be on or after from date")
	}
	if filter.DueFrom != nil && filter.DueTo != nil && filter.DueTo.Before(*filter.DueFrom) {
		return responses.BadRequest(c, "due_to must be on or after due_from")
	}

	if due := normalizeDueFilterParam(c.QueryParam("due")); due != "" {
		filter.DueAlert = due
	}

	docs, total, err := h.listDocsUC.Execute(c.Request().Context(), userID, filter)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}

	return responses.OKPaginated(c, "documents retrieved successfully", pagination.CreateResponse(docs, params, total))
}

// GetDocument handles GET /api/v1/archive/documents/:id
func (h *ArchiveHandler) GetDocument(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	doc, err := h.getDocUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), c.Param("id"))
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, doc, "document retrieved successfully")
}

// CreateDocument handles POST /api/v1/archive/documents
func (h *ArchiveHandler) CreateDocument(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	var req dtos.CreateDocumentRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceIDParam(c)
	}
	doc, err := h.createDocUC.Execute(c.Request().Context(), userID, &req)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.Created(c, doc, "document created successfully")
}

// UpdateDocument handles PATCH /api/v1/archive/documents/:id
func (h *ArchiveHandler) UpdateDocument(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	var req dtos.UpdateDocumentRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}
	wsID := workspaceIDParam(c)
	if wsID == "" {
		wsID = req.WorkspaceID
	}
	doc, err := h.updateDocUC.Execute(c.Request().Context(), userID, wsID, c.Param("id"), &req)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, doc, "document updated successfully")
}

// ArchiveDocument handles POST /api/v1/archive/documents/:id/archive
func (h *ArchiveHandler) ArchiveDocument(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	doc, err := h.archiveDocUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), c.Param("id"))
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, doc, "document archived successfully")
}

// ListCategories handles GET /api/v1/archive/categories
func (h *ArchiveHandler) ListCategories(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	cats, err := h.listCatsUC.Execute(c.Request().Context(), userID, workspaceIDParam(c))
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, cats, "categories retrieved successfully")
}

// CreateCategory handles POST /api/v1/archive/categories
func (h *ArchiveHandler) CreateCategory(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	var req dtos.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceIDParam(c)
	}
	cat, err := h.createCatUC.Execute(c.Request().Context(), userID, &req)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.Created(c, cat, "category created successfully")
}

// UpdateCategory handles PATCH /api/v1/archive/categories/:code
func (h *ArchiveHandler) UpdateCategory(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	var req dtos.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}
	wsID := workspaceIDParam(c)
	if wsID == "" {
		wsID = req.WorkspaceID
	}
	cat, err := h.updateCatUC.Execute(c.Request().Context(), userID, wsID, c.Param("code"), &req, currentUserIsSuperAdmin(c))
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, cat, "category updated successfully")
}

// DeactivateCategory handles DELETE /api/v1/archive/categories/:code
func (h *ArchiveHandler) DeactivateCategory(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	if err := h.deactivateCatUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), c.Param("code")); err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, nil, "category deactivated successfully")
}

func mapArchiveAPIError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domainerrors.ErrUserIDRequired),
		errors.Is(err, domainerrors.ErrTitleRequired),
		errors.Is(err, domainerrors.ErrCategoryRequired),
		errors.Is(err, domainerrors.ErrInvalidCategory),
		errors.Is(err, domainerrors.ErrDocumentIDRequired),
		errors.Is(err, domainerrors.ErrFileIDRequired),
		errors.Is(err, domainerrors.ErrFileRequired),
		errors.Is(err, domainerrors.ErrFileTooLarge),
		errors.Is(err, domainerrors.ErrInvalidContentType),
		errors.Is(err, domainerrors.ErrTooManyFiles),
		errors.Is(err, domainerrors.ErrWorkspaceNameRequired),
		errors.Is(err, domainerrors.ErrWorkspaceIDRequired),
		errors.Is(err, domainerrors.ErrEmailRequired),
		errors.Is(err, domainerrors.ErrInvalidMemberRole),
		errors.Is(err, domainerrors.ErrCannotInviteSelf),
		errors.Is(err, domainerrors.ErrAlreadyMember),
		errors.Is(err, domainerrors.ErrCannotModifyOwner),
		errors.Is(err, domainerrors.ErrHouseholdOnlyInvite),
		errors.Is(err, domainerrors.ErrTooManyExtraFields),
		errors.Is(err, domainerrors.ErrInvalidExtraField),
		errors.Is(err, domainerrors.ErrCategoryLabelRequired),
		errors.Is(err, domainerrors.ErrCategoryLabelTooLong),
		errors.Is(err, domainerrors.ErrCategoryDuplicateLabel),
		errors.Is(err, domainerrors.ErrTooManyCustomCategories),
		errors.Is(err, domainerrors.ErrCannotModifySystemCategory),
		errors.Is(err, domainerrors.ErrCategoryInUse):
		return responses.BadRequest(c, err.Error())
	case errors.Is(err, domainerrors.ErrDocumentNotFound),
		errors.Is(err, domainerrors.ErrWorkspaceNotFound),
		errors.Is(err, domainerrors.ErrFileNotFound),
		errors.Is(err, domainerrors.ErrInviteeNotFound),
		errors.Is(err, domainerrors.ErrCategoryNotFound):
		return responses.NotFound(c, err.Error())
	case errors.Is(err, domainerrors.ErrNotWorkspaceMember),
		errors.Is(err, domainerrors.ErrInsufficientWorkspaceRole):
		return responses.Forbidden(c, err.Error())
	default:
		return err
	}
}

func workspaceIDParam(c echo.Context) string {
	if id := strings.TrimSpace(c.QueryParam("workspace_id")); id != "" {
		return id
	}
	return strings.TrimSpace(c.FormValue("workspace_id"))
}

func currentUserID(c echo.Context) string {
	if u, ok := c.Get("user").(*authentities.User); ok && u != nil {
		return u.ID
	}
	if id := c.Request().Header.Get("X-User-ID"); id != "" {
		return id
	}
	return ""
}

func currentUserIsSuperAdmin(c echo.Context) bool {
	u, ok := c.Get("user").(*authentities.User)
	if !ok || u == nil {
		return false
	}
	for i := range u.Roles {
		if u.Roles[i].Name == "super_admin" {
			return true
		}
	}
	return false
}

func parseOptionalDate(raw string) (*time.Time, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, false, err
	}
	return &t, true, nil
}

func parseOptionalAmountCents(raw string) (*int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	raw = strings.ReplaceAll(raw, ",", ".")
	// Accept pesos with optional decimals; store cents.
	if strings.Contains(raw, ".") {
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, err
		}
		cents := int64(f*100 + 0.5) //nolint:mnd
		return &cents, nil
	}
	pesos, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	cents := pesos * 100 //nolint:mnd
	return &cents, nil
}
