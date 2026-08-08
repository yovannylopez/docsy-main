package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

type stubEnsureUC struct {
	ws *dtos.WorkspaceResponse
}

func (s stubEnsureUC) Execute(_ context.Context, _ string) (*dtos.WorkspaceResponse, error) {
	return s.ws, nil
}

type stubListDocsUC struct {
	docs  []dtos.DocumentResponse
	total int
}

func (s stubListDocsUC) Execute(_ context.Context, _ string, _ dtos.ListDocumentsFilter) ([]dtos.DocumentResponse, int, error) {
	return s.docs, s.total, nil
}

type stubGetDocUC struct {
	doc *dtos.DocumentResponse
	err error
}

func (s stubGetDocUC) Execute(_ context.Context, _, _, _ string) (*dtos.DocumentResponse, error) {
	return s.doc, s.err
}

type stubCreateDocUC struct {
	doc *dtos.DocumentResponse
}

func (s stubCreateDocUC) Execute(_ context.Context, _ string, _ *dtos.CreateDocumentRequest) (*dtos.DocumentResponse, error) {
	return s.doc, nil
}

type stubCreateWithFileUC struct {
	doc  *dtos.DocumentResponse
	file *dtos.DocumentFileResponse
	err  error
}

func (s stubCreateWithFileUC) Execute(
	_ context.Context,
	_ string,
	_ *dtos.CreateDocumentRequest,
	_, _ string,
	_ []byte,
) (*dtos.DocumentResponse, *dtos.DocumentFileResponse, error) {
	return s.doc, s.file, s.err
}

type stubUpdateDocUC struct {
	doc *dtos.DocumentResponse
}

func (s stubUpdateDocUC) Execute(_ context.Context, _, _, _ string, _ *dtos.UpdateDocumentRequest) (*dtos.DocumentResponse, error) {
	return s.doc, nil
}

type stubArchiveDocUC struct {
	doc *dtos.DocumentResponse
}

func (s stubArchiveDocUC) Execute(_ context.Context, _, _, _ string) (*dtos.DocumentResponse, error) {
	return s.doc, nil
}

type stubListCatsUC struct {
	cats []dtos.DocumentCategoryResponse
}

func (s stubListCatsUC) Execute(_ context.Context) ([]dtos.DocumentCategoryResponse, error) {
	return s.cats, nil
}

type stubListFoldersUC struct {
	folders []dtos.CategoryFolderResponse
	err     error
}

func (s stubListFoldersUC) Execute(_ context.Context, _, _, _ string) ([]dtos.CategoryFolderResponse, error) {
	return s.folders, s.err
}

type stubListWorkspacesUC struct {
	workspaces []dtos.WorkspaceResponse
}

func (s stubListWorkspacesUC) Execute(_ context.Context, _ string) ([]dtos.WorkspaceResponse, error) {
	return s.workspaces, nil
}

type stubCreateHouseholdUC struct {
	ws *dtos.WorkspaceResponse
}

func (s stubCreateHouseholdUC) Execute(_ context.Context, _ string, _ *dtos.CreateHouseholdRequest) (*dtos.WorkspaceResponse, error) {
	return s.ws, nil
}

type stubListMembersUC struct {
	members []dtos.WorkspaceMemberResponse
}

func (s stubListMembersUC) Execute(_ context.Context, _, _ string) ([]dtos.WorkspaceMemberResponse, error) {
	return s.members, nil
}

type stubInviteMemberUC struct {
	member *dtos.WorkspaceMemberResponse
}

func (s stubInviteMemberUC) Execute(_ context.Context, _, _ string, _ *dtos.InviteMemberRequest) (*dtos.WorkspaceMemberResponse, error) {
	return s.member, nil
}

type stubUpdateMemberRoleUC struct {
	member *dtos.WorkspaceMemberResponse
}

func (s stubUpdateMemberRoleUC) Execute(_ context.Context, _, _, _, _ string) (*dtos.WorkspaceMemberResponse, error) {
	return s.member, nil
}

type stubRemoveMemberUC struct{}

func (s stubRemoveMemberUC) Execute(_ context.Context, _, _, _ string) error {
	return nil
}

func newTestArchiveHandler() *ArchiveHandler {
	now := time.Now().UTC()
	return NewArchiveHandler(
		stubEnsureUC{ws: &dtos.WorkspaceResponse{ID: "ws-1", Name: "Mi archivo", Type: "personal", MemberRole: "owner"}},
		stubListWorkspacesUC{workspaces: []dtos.WorkspaceResponse{{ID: "ws-1", Name: "Mi archivo", Type: "personal", MemberRole: "owner"}}},
		stubCreateHouseholdUC{ws: &dtos.WorkspaceResponse{ID: "ws-h", Name: "Hogar", Type: "household", MemberRole: "owner"}},
		stubListMembersUC{},
		stubInviteMemberUC{},
		stubUpdateMemberRoleUC{},
		stubRemoveMemberUC{},
		stubListDocsUC{docs: []dtos.DocumentResponse{{ID: "d-1", Title: "Gas", Status: "active", CreatedAt: now, UpdatedAt: now}}, total: 1},
		stubGetDocUC{doc: &dtos.DocumentResponse{ID: "d-1", Title: "Gas"}},
		stubCreateDocUC{doc: &dtos.DocumentResponse{ID: "d-2", Title: "Nuevo"}},
		stubUpdateDocUC{doc: &dtos.DocumentResponse{ID: "d-1", Title: "Gas editado"}},
		stubArchiveDocUC{doc: &dtos.DocumentResponse{ID: "d-1", Status: "archived"}},
		stubListCatsUC{cats: []dtos.DocumentCategoryResponse{{Code: "utilities", LabelES: "Servicios públicos"}}},
		stubUploadFileUC{file: &dtos.DocumentFileResponse{ID: "f-1", OriginalName: "a.pdf"}},
		stubListFilesUC{files: []dtos.DocumentFileResponse{{ID: "f-1", OriginalName: "a.pdf", ContentType: "application/pdf"}}},
		stubDownloadFileUC{result: &dtos.DownloadDocumentFileResult{
			File: &dtos.DocumentFileResponse{ID: "f-1", OriginalName: "a.pdf", ContentType: "application/pdf"},
			Data: []byte("%PDF"),
		}},
		stubDeleteFileUC{},
	)
}

type stubUploadFileUC struct {
	file *dtos.DocumentFileResponse
	err  error
}

func (s stubUploadFileUC) Execute(_ context.Context, _, _, _, _, _ string, _ []byte) (*dtos.DocumentFileResponse, error) {
	return s.file, s.err
}

type stubListFilesUC struct {
	files []dtos.DocumentFileResponse
	err   error
}

func (s stubListFilesUC) Execute(_ context.Context, _, _, _ string) ([]dtos.DocumentFileResponse, error) {
	return s.files, s.err
}

type stubDownloadFileUC struct {
	result *dtos.DownloadDocumentFileResult
	err    error
}

func (s stubDownloadFileUC) Execute(_ context.Context, _, _, _, _ string) (*dtos.DownloadDocumentFileResult, error) {
	return s.result, s.err
}

type stubDeleteFileUC struct {
	err error
}

func (s stubDeleteFileUC) Execute(_ context.Context, _, _, _, _ string) error {
	return s.err
}

type stubSuggestOCRUC struct {
	resp *dtos.OCRSuggestionResponse
	err  error
}

func (s stubSuggestOCRUC) Execute(
	_ context.Context, _, _, _ string, _ []byte,
) (*dtos.OCRSuggestionResponse, error) {
	return s.resp, s.err
}

func withUser(c echo.Context) {
	c.Set("user", &authentities.User{ID: "user-1"})
}

func TestArchiveHandler_GetMyWorkspace(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/archive/workspaces/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	withUser(c)

	err := newTestArchiveHandler().GetMyWorkspace(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestArchiveHandler_ListDocuments(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/archive/documents?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	withUser(c)

	err := newTestArchiveHandler().ListDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "documents retrieved successfully")
}

func TestArchiveHandler_ListCategories(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/archive/categories", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := newTestArchiveHandler().ListCategories(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestArchiveHandler_UnauthorizedWithoutUser(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/archive/documents", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := newTestArchiveHandler().ListDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestArchiveHandler_CreateDocument(t *testing.T) {
	e := echo.New()
	body := `{"category_code":"taxes","title":"Predial"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/archive/documents", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	withUser(c)

	err := newTestArchiveHandler().CreateDocument(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "document created successfully")
}

func TestArchiveHandler_ArchiveDocument(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/archive/documents/d-1/archive", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d-1")
	withUser(c)

	err := newTestArchiveHandler().ArchiveDocument(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "document archived successfully")
}

func TestFormatAmountPesos(t *testing.T) {
	assert.Equal(t, "1250.50", formatAmountPesos(125050))
	assert.Equal(t, "0.01", formatAmountPesos(1))
}
