package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/templates"
)

func newTestPageHandler() *ArchivePageHandler {
	now := time.Now().UTC()
	ws := &dtos.WorkspaceResponse{
		ID: "ws-1", Name: "Mi archivo", Type: "personal", OwnerUserID: "u1",
		IsActive: true, MemberRole: "owner", CreatedAt: now, UpdatedAt: now,
	}
	return NewArchivePageHandler(
		stubEnsureUC{ws: ws},
		stubListWorkspacesUC{workspaces: []dtos.WorkspaceResponse{*ws}},
		stubCreateHouseholdUC{ws: &dtos.WorkspaceResponse{ID: "ws-h", Name: "Hogar", Type: "household", MemberRole: "owner"}},
		stubListMembersUC{},
		stubInviteMemberUC{},
		stubRemoveMemberUC{},
		stubListDocsUC{docs: []dtos.DocumentResponse{{
			ID: "d-1", Title: "Gas", CategoryLabel: "Servicios públicos", Status: "active",
			PrimaryOriginalName: "gas.pdf", PrimaryContentType: "application/pdf",
			CreatedAt: now, UpdatedAt: now,
		}}, total: 1},
		stubListFoldersUC{folders: []dtos.CategoryFolderResponse{
			{Code: categoryCodeUtilities, LabelES: "Servicios públicos", Count: 1},
			{Code: "taxes", LabelES: "Impuestos", Count: 0},
		}},
		stubGetDocUC{doc: &dtos.DocumentResponse{ID: "d-1", Title: "Gas", CategoryCode: categoryCodeUtilities}},
		stubCreateDocUC{doc: &dtos.DocumentResponse{ID: "d-2", Title: "Nuevo"}},
		stubCreateWithFileUC{
			doc:  &dtos.DocumentResponse{ID: "d-2", Title: "Nuevo"},
			file: &dtos.DocumentFileResponse{ID: "f-2", OriginalName: "nuevo.pdf"},
		},
		stubUpdateDocUC{doc: &dtos.DocumentResponse{ID: "d-1", Title: "Gas editado"}},
		stubListCatsUC{cats: []dtos.DocumentCategoryResponse{{Code: categoryCodeUtilities, LabelES: "Servicios públicos"}}},
		stubCreateCatUC{cat: &dtos.DocumentCategoryResponse{Code: "c_1", LabelES: "Custom"}},
		stubUpdateCatUC{cat: &dtos.DocumentCategoryResponse{Code: "c_1", LabelES: "Custom 2"}},
		stubDeactivateCatUC{},
		stubUploadFileUC{file: &dtos.DocumentFileResponse{ID: "f-1", OriginalName: "a.pdf"}},
		stubListFilesUC{files: []dtos.DocumentFileResponse{{
			ID: "f-1", OriginalName: "factura.pdf", ContentType: "application/pdf", SizeBytes: 1024, UploadedAt: now,
		}}},
		stubDownloadFileUC{result: &dtos.DownloadDocumentFileResult{
			File: &dtos.DocumentFileResponse{ID: "f-1", OriginalName: "factura.pdf", ContentType: "application/pdf"},
			Data: []byte("%PDF"),
		}},
		stubDeleteFileUC{},
		stubSuggestOCRUC{resp: &dtos.OCRSuggestionResponse{Title: "Sugerido", Amount: "1000"}},
	)
}

func renderPage(t *testing.T, path string, handle func(echo.Context) error) *httptest.ResponseRecorder {
	t.Helper()
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)

	e := echo.New()
	e.Renderer = renderer
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "u1", Email: "a@b.com", FirstName: "Ana"})

	err = handle(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	return rec
}

func TestArchivePageHandler_ShowArchive_RedirectsToPersonalDocs(t *testing.T) {
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)

	h := newTestPageHandler()
	e := echo.New()
	e.Renderer = renderer
	req := httptest.NewRequest(http.MethodGet, "/archivo", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "u1", Email: "a@b.com", FirstName: "Ana"})

	err = h.ShowArchive(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/archivo/documentos?workspace_id=ws-1")
}

func TestArchivePageHandler_ShowArchive_Hub(t *testing.T) {
	h := newTestPageHandler()
	rec := renderPage(t, "/archivo?hub=1", h.ShowArchive)
	assert.Contains(t, rec.Body.String(), "Mi archivo")
	assert.Contains(t, rec.Body.String(), "Crear hogar")
	assert.Contains(t, rec.Body.String(), "/archivo/documentos")
}

func TestArchivePageHandler_ListDocuments_Folders(t *testing.T) {
	h := newTestPageHandler()
	rec := renderPage(t, "/archivo/documentos", h.ListDocuments)
	assert.Contains(t, rec.Body.String(), "folder-card")
	assert.Contains(t, rec.Body.String(), "Servicios públicos")
	assert.Contains(t, rec.Body.String(), "Impuestos")
	assert.Contains(t, rec.Body.String(), "Nuevo")
}

func TestArchivePageHandler_ListDocuments_InCategory(t *testing.T) {
	h := newTestPageHandler()
	rec := renderPage(t, "/archivo/documentos?category=utilities&limit=10&offset=0", h.ListDocuments)
	assert.Contains(t, rec.Body.String(), "Gas")
	assert.Contains(t, rec.Body.String(), "files-grid")
	assert.Contains(t, rec.Body.String(), "Servicios públicos")
	assert.Contains(t, rec.Body.String(), "file-ext-badge")
	assert.Contains(t, rec.Body.String(), "PDF")
}

func TestArchivePageHandler_ShowCreate(t *testing.T) {
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)

	h := newTestPageHandler()
	e := echo.New()
	e.Renderer = renderer
	req := httptest.NewRequest(http.MethodGet, "/archivo/documentos/nuevo", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "u1", Email: "a@b.com"})

	err = h.ShowCreate(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "category_code")
	assert.Contains(t, rec.Body.String(), "Servicios públicos")
	assert.Contains(t, rec.Body.String(), "Analizar con OCR")
}

func TestArchivePageHandler_SuggestOCR_OK(t *testing.T) {
	h := newTestPageHandler()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "a.png")
	require.NoError(t, err)
	_, err = part.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	require.NoError(t, err)
	require.NoError(t, w.Close())

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/archivo/documentos/ocr-sugerir", &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "u1"})

	err = h.SuggestOCR(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Sugerido")
}

func TestArchivePageHandler_SuggestOCR_MissingFile(t *testing.T) {
	h := newTestPageHandler()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/archivo/documentos/ocr-sugerir", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "u1"})

	err := h.SuggestOCR(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestArchivePageHandler_SuggestOCR_Unavailable(t *testing.T) {
	h := newTestPageHandler()
	h.suggestOCRUC = stubSuggestOCRUC{err: domainerrors.ErrOCRUnavailable}
	e := echo.New()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "a.png")
	require.NoError(t, err)
	_, _ = part.Write([]byte{0x89, 0x50, 0x4E, 0x47})
	require.NoError(t, w.Close())
	req := httptest.NewRequest(http.MethodPost, "/archivo/documentos/ocr-sugerir", &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &authentities.User{ID: "u1"})

	err = h.SuggestOCR(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "OCR")
}

func TestPreviewKind(t *testing.T) {
	assert.Equal(t, previewKindPDF, previewKind("application/pdf", "a.pdf"))
	assert.Equal(t, previewKindPDF, previewKind("", "doc.PDF"))
	assert.Equal(t, previewKindImage, previewKind("image/png", "x.png"))
	assert.Equal(t, previewKindImage, previewKind("", "photo.jpeg"))
	assert.Equal(t, previewKindOther, previewKind("application/octet-stream", "bin.dat"))
}

func TestArchivePageHandler_DownloadDocumentFile_Inline(t *testing.T) {
	h := newTestPageHandler()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/archivo/documentos/d-1/archivos/f-1?inline=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "fileId")
	c.SetParamValues("d-1", "f-1")
	c.Set("user", &authentities.User{ID: "u1"})

	err := h.DownloadDocumentFile(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "inline")
	assert.Equal(t, "application/pdf", rec.Header().Get(echo.HeaderContentType))
}

func TestArchivePageHandler_ShowEdit_IncludesPreviewControls(t *testing.T) {
	renderer, err := templates.NewRenderer()
	require.NoError(t, err)

	h := newTestPageHandler()
	e := echo.New()
	e.Renderer = renderer
	req := httptest.NewRequest(http.MethodGet, "/archivo/documentos/d-1/editar", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d-1")
	c.Set("user", &authentities.User{ID: "u1", Email: "a@b.com"})

	err = h.ShowEdit(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Ver aquí")
	assert.Contains(t, rec.Body.String(), "inline=1")
	assert.Contains(t, rec.Body.String(), "document-preview-modal")
	assert.Contains(t, rec.Body.String(), "section-heading")
}
