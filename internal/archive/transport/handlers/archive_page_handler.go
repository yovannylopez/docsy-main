package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

const (
	archiveHomeTitle         = "Mi archivo"
	archiveDocsTitle         = "Documentos"
	archiveDocsSubtitle      = "Facturas, servicios, impuestos y más"
	archiveCreateTitle       = "Nuevo documento"
	archiveEditTitle         = "Editar documento"
	archiveHouseholdTitle    = "Nuevo hogar"
	archiveMembersTitle      = "Miembros del hogar"
	archiveHouseholdSubtitle = "Archivo compartido con tu familia"
	msgDocsLoadError         = "No se pudieron cargar los documentos. Intenta de nuevo."
	msgDocCreateError        = "No se pudo crear el documento. Verifica los datos e intenta de nuevo."
	msgDocUpdateError        = "No se pudo actualizar el documento. Verifica los datos e intenta de nuevo."
	msgDocNotFound           = "El documento solicitado no existe."
	msgDocCreated            = "Documento creado correctamente."
	msgDocUpdated            = "Documento actualizado correctamente."
	msgFileUploaded          = "Archivo subido correctamente."
	msgFileDeleted           = "Archivo eliminado correctamente."
	msgFileUploadError       = "No se pudo subir el archivo. Verifica el tipo y el tamaño."
	msgFileRequiredCreate    = "Debes adjuntar un archivo para crear el documento."
	msgCannotDeleteLastFile  = "El documento debe conservar al menos un archivo adjunto."
	msgOCRUnavailable        = "El análisis OCR no está disponible. Instala Tesseract o actívalo en la configuración."
	msgOCRUnsupported        = "Este tipo de archivo no admite OCR. Usa PDF o imagen (JPG, PNG, TIFF, WebP, GIF)."
	msgOCRNoText             = "No se pudo leer texto del archivo. Prueba con una imagen más nítida o un PDF con texto."
	msgOCRFailed             = "No se pudo analizar el archivo con OCR. Intenta de nuevo."
	msgInvalidAmount         = "el monto no es válido"
	msgInvalidDate           = "hay una fecha con formato inválido (usa AAAA-MM-DD)"
	msgWorkspacesLoadError   = "No se pudieron cargar tus archivos. Intenta de nuevo."
	msgHouseholdCreateError  = "No se pudo crear el hogar. Verifica los datos e intenta de nuevo."
	msgHouseholdCreated      = "Hogar creado correctamente."
	msgMembersLoadError      = "No se pudieron cargar los miembros. Intenta de nuevo."
	msgMemberInvited         = "Miembro invitado correctamente."
	msgMemberRemoved         = "Miembro eliminado correctamente."
	msgMemberInviteError     = "No se pudo invitar al miembro. Verifica el correo e intenta de nuevo."
	msgMemberRemoveError     = "No se pudo eliminar al miembro. Intenta de nuevo."
	defaultDocsPageLimit     = 10
	defaultFormCategory      = "other"
	docsModeFolders          = "folders"
	docsModeDocuments        = "documents"
)

// ArchivePageData holds view data for /archivo.
type ArchivePageData struct {
	weblayout.AppLayoutData
	Personal   *WorkspaceListItem
	Households []WorkspaceListItem
	Error      string
}

// WorkspaceListItem is a row on the archive home page.
type WorkspaceListItem struct {
	ID          string
	Name        string
	Type        string
	TypeLabel   string
	MemberRole  string
	DocsURL     string
	MembersURL  string
	ShowMembers bool
}

// HouseholdFormPageData holds create household form views.
type HouseholdFormPageData struct {
	weblayout.AppLayoutData
	Form    HouseholdForm
	Error   string
	Success string
}

// HouseholdForm holds HTML form values for a new household.
type HouseholdForm struct {
	Name         string
	GeneralError string
}

// MembersPageData holds the household members view.
type MembersPageData struct {
	weblayout.AppLayoutData
	Workspace  *dtos.WorkspaceResponse
	Members    []dtos.WorkspaceMemberResponse
	CanManage  bool
	InviteForm InviteMemberForm
	Error      string
	Success    string
}

// InviteMemberForm holds invite form values.
type InviteMemberForm struct {
	Email string
	Role  string
}

// DocumentsListPageData holds the documents list view.
type DocumentsListPageData struct {
	weblayout.AppLayoutData
	WorkspaceID   string
	Workspaces    []dtos.WorkspaceResponse
	Documents     []DocumentListItem
	Folders       []CategoryFolderItem
	Total         int
	Query         string
	Category      string
	CategoryLabel string
	Status        string
	View          string // grid | list
	Mode          string // folders | documents
	BrowseURL     string // folders root (no category/query)
	CreateURL     string // new document (optionally preselects category)
	Categories    []dtos.DocumentCategoryResponse
	Error         string
	Success       string
	Pagination    weblayout.PaginationData
}

// CategoryFolderItem is a virtual folder card for the documents browser.
type CategoryFolderItem struct {
	Code     string
	Label    string
	Count    int
	OpenURL  string
	IconTone string
	MetaLine string
}

// DocumentListItem is a card/row for the documents browser.
type DocumentListItem struct {
	ID              string
	Title           string
	CategoryCode    string
	CategoryLabel   string
	Issuer          string
	ReferenceNumber string
	DocumentDate    string
	DueDate         string
	AmountDisplay   string
	Status          string
	EditURL         string
	IconTone        string // CSS modifier: tone-red, tone-blue, ...
	MetaLine        string // secondary line under title
	FileExtension   string // PDF, PNG, ... empty when no attachment
	FileBadgeClass  string // CSS modifier: ext-pdf, ext-image, ext-other
	HasAttachment   bool
}

// DocumentFormPageData holds create/edit form views.
type DocumentFormPageData struct {
	weblayout.AppLayoutData
	WorkspaceID string
	Form        DocumentForm
	Categories  []dtos.DocumentCategoryResponse
	Files       []DocumentFileItem
	Error       string
	Success     string
	IsEdit      bool
}

// DocumentFileItem is a row for attachments on the edit form.
type DocumentFileItem struct {
	ID            string
	OriginalName  string
	ContentType   string
	SizeDisplay   string
	UploadedAt    string
	DownloadURL   string
	PreviewURL    string
	DeleteURL     string
	Extension     string
	BadgeColor    string
	PreviewKind   string // pdf | image | other
	IsPreviewable bool
}

// DocumentForm holds HTML form values.
type DocumentForm struct {
	ID              string
	WorkspaceID     string
	CategoryCode    string
	Title           string
	DocumentDate    string
	DueDate         string
	Issuer          string
	ReferenceNumber string
	Amount          string
	Currency        string
	Notes           string
	ExtraFields     []dtos.ExtraFieldDTO
	ExtraFieldsJSON string
	GeneralError    string
}

// ArchivePageHandler serves HTMX/HTML archive pages.
type ArchivePageHandler struct {
	ensurePersonalUC  ports.EnsurePersonalWorkspaceService
	listWorkspacesUC  ports.ListWorkspacesService
	createHouseholdUC ports.CreateHouseholdWorkspaceService
	listMembersUC     ports.ListWorkspaceMembersService
	inviteMemberUC    ports.InviteWorkspaceMemberService
	removeMemberUC    ports.RemoveWorkspaceMemberService
	listDocsUC        ports.ListDocumentsService
	listFoldersUC     ports.ListCategoryFoldersService
	getDocUC          ports.GetDocumentService
	createDocUC       ports.CreateDocumentService
	createWithFileUC  ports.CreateDocumentWithFileService
	updateDocUC       ports.UpdateDocumentService
	listCatsUC        ports.ListCategoriesService
	createCatUC       ports.CreateCategoryService
	updateCatUC       ports.UpdateCategoryService
	deactivateCatUC   ports.DeactivateCategoryService
	uploadFileUC      ports.UploadDocumentFileService
	listFilesUC       ports.ListDocumentFilesService
	downloadFileUC    ports.DownloadDocumentFileService
	deleteFileUC      ports.DeleteDocumentFileService
	suggestOCRUC      ports.SuggestDocumentFieldsService
}

// NewArchivePageHandler creates the page handler.
func NewArchivePageHandler(
	ensurePersonalUC ports.EnsurePersonalWorkspaceService,
	listWorkspacesUC ports.ListWorkspacesService,
	createHouseholdUC ports.CreateHouseholdWorkspaceService,
	listMembersUC ports.ListWorkspaceMembersService,
	inviteMemberUC ports.InviteWorkspaceMemberService,
	removeMemberUC ports.RemoveWorkspaceMemberService,
	listDocsUC ports.ListDocumentsService,
	listFoldersUC ports.ListCategoryFoldersService,
	getDocUC ports.GetDocumentService,
	createDocUC ports.CreateDocumentService,
	createWithFileUC ports.CreateDocumentWithFileService,
	updateDocUC ports.UpdateDocumentService,
	listCatsUC ports.ListCategoriesService,
	createCatUC ports.CreateCategoryService,
	updateCatUC ports.UpdateCategoryService,
	deactivateCatUC ports.DeactivateCategoryService,
	uploadFileUC ports.UploadDocumentFileService,
	listFilesUC ports.ListDocumentFilesService,
	downloadFileUC ports.DownloadDocumentFileService,
	deleteFileUC ports.DeleteDocumentFileService,
	suggestOCRUC ports.SuggestDocumentFieldsService,
) *ArchivePageHandler {
	return &ArchivePageHandler{
		ensurePersonalUC:  ensurePersonalUC,
		listWorkspacesUC:  listWorkspacesUC,
		createHouseholdUC: createHouseholdUC,
		listMembersUC:     listMembersUC,
		inviteMemberUC:    inviteMemberUC,
		removeMemberUC:    removeMemberUC,
		listDocsUC:        listDocsUC,
		listFoldersUC:     listFoldersUC,
		getDocUC:          getDocUC,
		createDocUC:       createDocUC,
		createWithFileUC:  createWithFileUC,
		updateDocUC:       updateDocUC,
		listCatsUC:        listCatsUC,
		createCatUC:       createCatUC,
		updateCatUC:       updateCatUC,
		deactivateCatUC:   deactivateCatUC,
		uploadFileUC:      uploadFileUC,
		listFilesUC:       listFilesUC,
		downloadFileUC:    downloadFileUC,
		deleteFileUC:      deleteFileUC,
		suggestOCRUC:      suggestOCRUC,
	}
}

func (h *ArchivePageHandler) loadCategories(c echo.Context, workspaceID string) []dtos.DocumentCategoryResponse {
	cats, err := h.listCatsUC.Execute(c.Request().Context(), weblayout.CurrentUserID(c), workspaceID)
	if err != nil {
		return nil
	}
	return cats
}

// ShowArchive renders GET /archivo. With only a personal workspace, opens documents directly.
func (h *ArchivePageHandler) ShowArchive(c echo.Context) error {
	layout := weblayout.AppLayoutFromEcho(c, archiveHomeTitle, "Tu espacio para documentos personales y del hogar", "/archivo")

	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}

	workspaces, err := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "archive/home", ArchivePageData{
			AppLayoutData: layout,
			Error:         msgWorkspacesLoadError,
		})
	}

	var personal *WorkspaceListItem
	households := make([]WorkspaceListItem, 0, len(workspaces))
	for i := range workspaces {
		item := toWorkspaceListItem(&workspaces[i])
		if workspaces[i].Type == entities.WorkspaceTypePersonal && personal == nil {
			p := item
			personal = &p
			continue
		}
		if workspaces[i].Type == entities.WorkspaceTypeHousehold {
			households = append(households, item)
		}
	}

	// Default path: jump into personal documents (modern, zero friction).
	forceHub := strings.TrimSpace(c.QueryParam("hub")) == "1"
	if !forceHub && personal != nil && len(households) == 0 {
		return c.Redirect(http.StatusFound, personal.DocsURL)
	}

	return c.Render(http.StatusOK, "archive/home", ArchivePageData{
		AppLayoutData: layout,
		Personal:      personal,
		Households:    households,
	})
}

// ShowCreateHousehold renders GET /archivo/hogares/nuevo.
func (h *ArchivePageHandler) ShowCreateHousehold(c echo.Context) error {
	return c.Render(http.StatusOK, "archive/household_form", HouseholdFormPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveHouseholdTitle, archiveHouseholdSubtitle, "/archivo/hogares/nuevo"),
		Form:          HouseholdForm{Name: "Archivo del hogar"},
	})
}

// SubmitCreateHousehold handles POST /archivo/hogares/nuevo.
func (h *ArchivePageHandler) SubmitCreateHousehold(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}

	form := HouseholdForm{Name: strings.TrimSpace(c.FormValue("name"))}
	ws, err := h.createHouseholdUC.Execute(c.Request().Context(), userID, &dtos.CreateHouseholdRequest{Name: form.Name})
	if err != nil {
		form.GeneralError = mapHouseholdFormError(err, msgHouseholdCreateError)
		return c.Render(http.StatusUnprocessableEntity, "archive/household_form", HouseholdFormPageData{
			AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveHouseholdTitle, archiveHouseholdSubtitle, "/archivo/hogares/nuevo"),
			Form:          form,
			Error:         form.GeneralError,
		})
	}

	redirectURL := fmt.Sprintf("/archivo/hogares/%s/miembros?created=1", ws.ID)
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

// ShowMembers renders GET /archivo/hogares/:id/miembros.
func (h *ArchivePageHandler) ShowMembers(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	workspaceID := c.Param("id")
	layout := weblayout.AppLayoutFromEcho(c, archiveMembersTitle, "Administra quién puede ver y editar", "/archivo/hogares/"+workspaceID+"/miembros")

	workspaces, err := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "archive/members", MembersPageData{
			AppLayoutData: layout,
			Error:         msgMembersLoadError,
		})
	}
	var ws *dtos.WorkspaceResponse
	for i := range workspaces {
		if workspaces[i].ID == workspaceID {
			ws = &workspaces[i]
			break
		}
	}
	if ws == nil {
		return c.Render(http.StatusNotFound, "forbidden", weblayout.AppLayoutFromEcho(
			c, "Hogar no encontrado", "El archivo solicitado no existe o no tienes acceso.", "/archivo",
		))
	}

	members, err := h.listMembersUC.Execute(c.Request().Context(), userID, workspaceID)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "archive/members", MembersPageData{
			AppLayoutData: layout,
			Workspace:     ws,
			Error:         msgMembersLoadError,
		})
	}

	success := ""
	switch {
	case c.QueryParam("created") == "1":
		success = msgHouseholdCreated
	case c.QueryParam("invited") == "1":
		success = msgMemberInvited
	case c.QueryParam("removed") == "1":
		success = msgMemberRemoved
	}

	return c.Render(http.StatusOK, "archive/members", MembersPageData{
		AppLayoutData: layout,
		Workspace:     ws,
		Members:       members,
		CanManage:     ws.MemberRole == entities.WorkspaceRoleOwner,
		InviteForm:    InviteMemberForm{Role: entities.WorkspaceRoleMember},
		Success:       success,
	})
}

// InviteMember handles POST /archivo/hogares/:id/miembros.
func (h *ArchivePageHandler) InviteMember(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	workspaceID := c.Param("id")
	inviteForm := InviteMemberForm{
		Email: strings.TrimSpace(c.FormValue("email")),
		Role:  strings.TrimSpace(c.FormValue("role")),
	}
	if inviteForm.Role == "" {
		inviteForm.Role = entities.WorkspaceRoleMember
	}

	_, err := h.inviteMemberUC.Execute(c.Request().Context(), userID, workspaceID, &dtos.InviteMemberRequest{
		Email: inviteForm.Email,
		Role:  inviteForm.Role,
	})
	if err != nil {
		return h.renderMembersWithError(c, workspaceID, inviteForm, mapMemberFormError(err, msgMemberInviteError))
	}

	redirectURL := fmt.Sprintf("/archivo/hogares/%s/miembros?invited=1", workspaceID)
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

// RemoveMember handles POST /archivo/hogares/:id/miembros/:userId/eliminar.
func (h *ArchivePageHandler) RemoveMember(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")

	if err := h.removeMemberUC.Execute(c.Request().Context(), userID, workspaceID, targetUserID); err != nil {
		return h.renderMembersWithError(c, workspaceID, InviteMemberForm{Role: entities.WorkspaceRoleMember}, mapMemberFormError(err, msgMemberRemoveError))
	}

	redirectURL := fmt.Sprintf("/archivo/hogares/%s/miembros?removed=1", workspaceID)
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

// ListDocuments renders GET /archivo/documentos.
func (h *ArchivePageHandler) ListDocuments(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}

	params, err := docsPaginationParser().ParseFromQuery(c.QueryParam("limit"), c.QueryParam("offset"))
	if err != nil {
		return h.renderDocsList(c, DocumentsListPageData{
			AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveDocsTitle, archiveDocsSubtitle, "/archivo/documentos"),
			Error:         err.Error(),
			Mode:          docsModeFolders,
		})
	}

	query := strings.TrimSpace(c.QueryParam("q"))
	category := strings.TrimSpace(c.QueryParam("category"))
	status := strings.TrimSpace(c.QueryParam("status"))
	view := strings.TrimSpace(strings.ToLower(c.QueryParam("view")))
	if view != "list" {
		view = "grid"
	}
	workspaceID := workspaceIDParam(c)
	if status == "" {
		status = entities.DocumentStatusActive
	}

	workspaces, _ := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
	cats := h.loadCategories(c, workspaceID)
	categoryLabel := categoryLabelFromList(cats, category)

	base := DocumentsListPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveDocsTitle, archiveDocsSubtitle, "/archivo/documentos"),
		WorkspaceID:   workspaceID,
		Workspaces:    workspaces,
		Query:         query,
		Category:      category,
		CategoryLabel: categoryLabel,
		Status:        status,
		View:          view,
		BrowseURL:     documentsBrowseURL(workspaceID, status),
		CreateURL:     documentsCreateURL(workspaceID, category),
		Categories:    cats,
		Success:       archiveFlash(c),
	}

	// Virtual folders home: no category selected and no search query.
	if category == "" && query == "" {
		folders, folderErr := h.listFoldersUC.Execute(c.Request().Context(), userID, workspaceID, status)
		if folderErr != nil {
			base.Mode = docsModeFolders
			base.Error = msgDocsLoadError
			return h.renderDocsList(c, base)
		}
		items := make([]CategoryFolderItem, 0, len(folders))
		totalDocs := 0
		for _, f := range folders {
			totalDocs += f.Count
			items = append(items, toFolderItem(f, workspaceID, status, view))
		}
		base.Mode = docsModeFolders
		base.Folders = items
		base.Total = totalDocs
		return h.renderDocsList(c, base)
	}

	docs, total, listErr := h.listDocsUC.Execute(c.Request().Context(), userID, dtos.ListDocumentsFilter{
		WorkspaceID: workspaceID,
		Category:    category,
		Query:       query,
		Status:      status,
		Limit:       params.Limit,
		Offset:      params.Offset,
	})
	if listErr != nil {
		base.Mode = docsModeDocuments
		base.Error = msgDocsLoadError
		return h.renderDocsList(c, base)
	}

	items := make([]DocumentListItem, 0, len(docs))
	for i := range docs {
		items = append(items, toListItem(&docs[i], workspaceID))
	}

	base.Mode = docsModeDocuments
	base.Documents = items
	base.Total = total
	base.Pagination = weblayout.NewPaginationData(params.Offset, params.Limit, total, "/archivo/documentos", c.QueryParams())
	return h.renderDocsList(c, base)
}

// ShowCreate renders GET /archivo/documentos/nuevo.
func (h *ArchivePageHandler) ShowCreate(c echo.Context) error {
	workspaceID := workspaceIDParam(c)
	cats := h.loadCategories(c, workspaceID)
	categoryCode := strings.TrimSpace(c.QueryParam("category_code"))
	if !categoryInList(cats, categoryCode) {
		categoryCode = defaultFormCategory
	}
	return c.Render(http.StatusOK, "archive/document_form", DocumentFormPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveCreateTitle, "Registra el documento y adjunta el archivo", "/archivo/documentos/nuevo"),
		WorkspaceID:   workspaceID,
		Form:          DocumentForm{Currency: entities.DefaultDocumentCurrency, CategoryCode: categoryCode, WorkspaceID: workspaceID},
		Categories:    cats,
		IsEdit:        false,
	})
}

// SubmitCreate handles POST /archivo/documentos/nuevo (multipart: metadata + required file).
func (h *ArchivePageHandler) SubmitCreate(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}

	form := bindDocumentForm(c, "")
	req, bindErr := formToCreateRequest(form)
	cats := h.loadCategories(c, form.WorkspaceID)
	if bindErr != nil {
		form.GeneralError = bindErr.Error()
		return h.renderForm(c, form, cats, false, bindErr.Error(), http.StatusUnprocessableEntity)
	}

	header, fileErr := c.FormFile("file")
	if fileErr != nil {
		form.GeneralError = msgFileRequiredCreate
		return h.renderForm(c, form, cats, false, msgFileRequiredCreate, http.StatusUnprocessableEntity)
	}
	src, err := header.Open()
	if err != nil {
		form.GeneralError = msgFileUploadError
		return h.renderForm(c, form, cats, false, msgFileUploadError, http.StatusUnprocessableEntity)
	}
	defer func() { _ = src.Close() }()

	data, err := io.ReadAll(src)
	if err != nil {
		form.GeneralError = msgFileUploadError
		return h.renderForm(c, form, cats, false, msgFileUploadError, http.StatusUnprocessableEntity)
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	_, _, err = h.createWithFileUC.Execute(
		c.Request().Context(), userID, req, header.Filename, contentType, data,
	)
	if err != nil {
		form.GeneralError = mapCreateWithFileError(err)
		return h.renderForm(c, form, cats, false, form.GeneralError, http.StatusUnprocessableEntity)
	}

	if weblayout.IsHTMXRequest(c) {
		redirect := "/archivo/documentos?created=1"
		if form.WorkspaceID != "" {
			redirect += "&workspace_id=" + form.WorkspaceID
		}
		c.Response().Header().Set("HX-Redirect", redirect)
		return c.NoContent(http.StatusOK)
	}
	redirect := "/archivo/documentos?created=1"
	if form.WorkspaceID != "" {
		redirect += "&workspace_id=" + form.WorkspaceID
	}
	return c.Redirect(http.StatusFound, redirect)
}

// ShowEdit renders GET /archivo/documentos/:id/editar.
func (h *ArchivePageHandler) ShowEdit(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	docID := c.Param("id")
	workspaceID := workspaceIDParam(c)
	doc, err := h.getDocUC.Execute(c.Request().Context(), userID, workspaceID, docID)
	cats := h.loadCategories(c, workspaceID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrDocumentNotFound) {
			return c.Render(http.StatusNotFound, "forbidden", weblayout.AppLayoutFromEcho(
				c, "Documento no encontrado", msgDocNotFound, "/archivo/documentos",
			))
		}
		return h.renderForm(c, DocumentForm{ID: docID, WorkspaceID: workspaceID}, cats, true, msgDocUpdateError, http.StatusInternalServerError)
	}

	form := documentToForm(doc)
	files := h.loadFileItems(c.Request().Context(), userID, doc.WorkspaceID, docID)
	success := ""
	switch {
	case c.QueryParam("updated") == "1":
		success = msgDocUpdated
	case c.QueryParam("uploaded") == "1":
		success = msgFileUploaded
	case c.QueryParam("deleted") == "1":
		success = msgFileDeleted
	}
	return c.Render(http.StatusOK, "archive/document_form", DocumentFormPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveEditTitle, "Actualiza los metadatos", "/archivo/documentos/"+docID+"/editar"),
		WorkspaceID:   doc.WorkspaceID,
		Form:          form,
		Categories:    cats,
		Files:         files,
		Success:       success,
		IsEdit:        true,
	})
}

// SubmitEdit handles POST /archivo/documentos/:id/editar.
func (h *ArchivePageHandler) SubmitEdit(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	docID := c.Param("id")
	form := bindDocumentForm(c, docID)
	req, bindErr := formToUpdateRequest(form)
	cats := h.loadCategories(c, form.WorkspaceID)
	if bindErr != nil {
		form.GeneralError = bindErr.Error()
		return h.renderForm(c, form, cats, true, bindErr.Error(), http.StatusUnprocessableEntity)
	}

	_, err := h.updateDocUC.Execute(c.Request().Context(), userID, form.WorkspaceID, docID, req)
	if err != nil {
		msg := mapDocFormError(err, msgDocUpdateError)
		form.GeneralError = msg
		return h.renderForm(c, form, cats, true, msg, http.StatusUnprocessableEntity)
	}

	redirectURL := fmt.Sprintf("/archivo/documentos/%s/editar?updated=1", docID)
	if form.WorkspaceID != "" {
		redirectURL += "&workspace_id=" + form.WorkspaceID
	}
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

func (h *ArchivePageHandler) renderDocsList(c echo.Context, data DocumentsListPageData) error {
	if weblayout.IsHTMXRequest(c) && strings.Contains(c.Request().Header.Get("HX-Target"), "archive-docs-table") {
		return c.Render(http.StatusOK, "partials/archive-docs-table", data)
	}
	return c.Render(http.StatusOK, "archive/documents", data)
}

func (h *ArchivePageHandler) renderForm(
	c echo.Context,
	form DocumentForm,
	cats []dtos.DocumentCategoryResponse,
	isEdit bool,
	errMsg string,
	status int,
) error {
	title := archiveCreateTitle
	subtitle := "Registra el documento y adjunta el archivo"
	route := "/archivo/documentos/nuevo"
	var files []DocumentFileItem
	if isEdit {
		title = archiveEditTitle
		subtitle = "Actualiza los metadatos"
		route = "/archivo/documentos/" + form.ID + "/editar"
		if userID := weblayout.CurrentUserID(c); userID != "" && form.ID != "" {
			files = h.loadFileItems(c.Request().Context(), userID, form.WorkspaceID, form.ID)
		}
	}
	return c.Render(status, "archive/document_form", DocumentFormPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, title, subtitle, route),
		WorkspaceID:   form.WorkspaceID,
		Form:          form,
		Categories:    cats,
		Files:         files,
		Error:         errMsg,
		IsEdit:        isEdit,
	})
}

func bindDocumentForm(c echo.Context, id string) DocumentForm {
	extras, extrasJSON := parseExtraFieldsForm(c.FormValue("extra_fields"))
	return DocumentForm{
		ID:              id,
		WorkspaceID:     workspaceIDParam(c),
		CategoryCode:    strings.TrimSpace(c.FormValue("category_code")),
		Title:           strings.TrimSpace(c.FormValue("title")),
		DocumentDate:    strings.TrimSpace(c.FormValue("document_date")),
		DueDate:         strings.TrimSpace(c.FormValue("due_date")),
		Issuer:          strings.TrimSpace(c.FormValue("issuer")),
		ReferenceNumber: strings.TrimSpace(c.FormValue("reference_number")),
		Amount:          strings.TrimSpace(c.FormValue("amount")),
		Currency:        strings.TrimSpace(c.FormValue("currency")),
		Notes:           strings.TrimSpace(c.FormValue("notes")),
		ExtraFields:     extras,
		ExtraFieldsJSON: extrasJSON,
	}
}

func parseExtraFieldsForm(raw string) ([]dtos.ExtraFieldDTO, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" || raw == "null" {
		return nil, "[]"
	}
	var fields []dtos.ExtraFieldDTO
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return nil, "[]"
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return fields, "[]"
	}
	return fields, string(b)
}

func marshalExtraFieldsJSON(fields []dtos.ExtraFieldDTO) string {
	if len(fields) == 0 {
		return "[]"
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func formToCreateRequest(form DocumentForm) (*dtos.CreateDocumentRequest, error) {
	docDate, err := parseFormDate(form.DocumentDate)
	if err != nil {
		return nil, errors.New(msgInvalidDate)
	}
	dueDate, err := parseFormDate(form.DueDate)
	if err != nil {
		return nil, errors.New(msgInvalidDate)
	}
	amount, err := parseOptionalAmountCents(form.Amount)
	if err != nil {
		return nil, errors.New(msgInvalidAmount)
	}
	currency := form.Currency
	if currency == "" {
		currency = entities.DefaultDocumentCurrency
	}
	return &dtos.CreateDocumentRequest{
		WorkspaceID:     form.WorkspaceID,
		CategoryCode:    form.CategoryCode,
		Title:           form.Title,
		DocumentDate:    docDate,
		DueDate:         dueDate,
		Issuer:          optionalNonEmpty(form.Issuer),
		ReferenceNumber: optionalNonEmpty(form.ReferenceNumber),
		AmountCents:     amount,
		Currency:        currency,
		Notes:           optionalNonEmpty(form.Notes),
		ExtraFields:     form.ExtraFields,
	}, nil
}

func formToUpdateRequest(form DocumentForm) (*dtos.UpdateDocumentRequest, error) {
	docDate, err := parseFormDate(form.DocumentDate)
	if err != nil {
		return nil, errors.New(msgInvalidDate)
	}
	dueDate, err := parseFormDate(form.DueDate)
	if err != nil {
		return nil, errors.New(msgInvalidDate)
	}
	amount, err := parseOptionalAmountCents(form.Amount)
	if err != nil {
		return nil, errors.New(msgInvalidAmount)
	}
	title := form.Title
	cat := form.CategoryCode
	currency := form.Currency
	if currency == "" {
		currency = entities.DefaultDocumentCurrency
	}
	issuer := form.Issuer
	ref := form.ReferenceNumber
	notes := form.Notes
	req := &dtos.UpdateDocumentRequest{
		WorkspaceID:     form.WorkspaceID,
		CategoryCode:    &cat,
		Title:           &title,
		DocumentDate:    docDate,
		DueDate:         dueDate,
		ClearDueDate:    form.DueDate == "",
		Issuer:          &issuer,
		ReferenceNumber: &ref,
		AmountCents:     amount,
		ClearAmount:     form.Amount == "",
		Currency:        &currency,
		Notes:           &notes,
		ExtraFields:     form.ExtraFields,
		SetExtraFields:  true,
	}
	return req, nil
}

func documentToForm(doc *dtos.DocumentResponse) DocumentForm {
	form := DocumentForm{
		ID:           doc.ID,
		WorkspaceID:  doc.WorkspaceID,
		CategoryCode: doc.CategoryCode,
		Title:        doc.Title,
		Currency:     doc.Currency,
	}
	if doc.DocumentDate != nil {
		form.DocumentDate = doc.DocumentDate.Format("2006-01-02")
	}
	if doc.DueDate != nil {
		form.DueDate = doc.DueDate.Format("2006-01-02")
	}
	if doc.Issuer != nil {
		form.Issuer = *doc.Issuer
	}
	if doc.ReferenceNumber != nil {
		form.ReferenceNumber = *doc.ReferenceNumber
	}
	if doc.AmountCents != nil {
		form.Amount = formatAmountPesos(*doc.AmountCents)
	}
	if doc.Notes != nil {
		form.Notes = *doc.Notes
	}
	form.ExtraFields = doc.ExtraFields
	form.ExtraFieldsJSON = marshalExtraFieldsJSON(doc.ExtraFields)
	return form
}

func toListItem(doc *dtos.DocumentResponse, workspaceID string) DocumentListItem {
	wsID := workspaceID
	if wsID == "" {
		wsID = doc.WorkspaceID
	}
	item := DocumentListItem{
		ID:            doc.ID,
		Title:         doc.Title,
		CategoryCode:  doc.CategoryCode,
		CategoryLabel: doc.CategoryLabel,
		Status:        doc.Status,
		EditURL:       documentEditURL(doc.ID, wsID),
		IconTone:      categoryIconTone(doc.CategoryCode),
	}
	if item.CategoryLabel == "" {
		item.CategoryLabel = doc.CategoryCode
	}
	if doc.Issuer != nil {
		item.Issuer = *doc.Issuer
	}
	if doc.ReferenceNumber != nil {
		item.ReferenceNumber = *doc.ReferenceNumber
	}
	if doc.DocumentDate != nil {
		item.DocumentDate = doc.DocumentDate.Format("02/01/2006")
	}
	if doc.DueDate != nil {
		item.DueDate = doc.DueDate.Format("02/01/2006")
	}
	if doc.AmountCents != nil {
		item.AmountDisplay = formatAmountPesos(*doc.AmountCents) + " " + doc.Currency
	}
	if name := strings.TrimSpace(doc.PrimaryOriginalName); name != "" {
		ext := fileExtension(name)
		item.FileExtension = ext
		item.FileBadgeClass = fileExtensionBadgeClass(ext, doc.PrimaryContentType, name)
		item.HasAttachment = true
	}
	item.MetaLine = documentMetaLine(item)
	return item
}

func fileExtensionBadgeClass(ext, contentType, filename string) string {
	switch strings.ToLower(ext) {
	case "pdf":
		return "ext-pdf"
	case "jpg", "jpeg", "png", "gif", "webp", "tif", "tiff":
		return "ext-image"
	case "xls", "xlsx":
		return "ext-sheet"
	case "doc", "docx":
		return "ext-word"
	default:
		switch previewKind(contentType, filename) {
		case previewKindPDF:
			return "ext-pdf"
		case previewKindImage:
			return "ext-image"
		default:
			return "ext-other"
		}
	}
}

const (
	categoryCodeIdentity      = "identity"
	categoryCodeHealth        = "health"
	categoryCodeFinance       = "finance"
	categoryCodeTaxes         = "taxes"
	categoryCodeProperty      = "property"
	categoryCodeInsurance     = "insurance"
	categoryCodeEducation     = "education"
	categoryCodeWork          = "work"
	categoryCodeLegal         = "legal"
	categoryCodeUtilities     = "utilities"
	categoryCodeInvoices      = "invoices"
	categoryCodePhotos        = "photos"
	documentMetaPartsCapacity = 3
)

func categoryIconTone(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case categoryCodeIdentity:
		return "tone-indigo"
	case categoryCodeHealth:
		return "tone-teal"
	case categoryCodeFinance:
		return "tone-green"
	case categoryCodeTaxes:
		return "tone-red"
	case categoryCodeProperty:
		return "tone-amber"
	case categoryCodeInsurance:
		return "tone-blue"
	case categoryCodeEducation:
		return "tone-violet"
	case categoryCodeWork:
		return "tone-indigo"
	case categoryCodeLegal:
		return "tone-red"
	case categoryCodeUtilities:
		return "tone-blue"
	case categoryCodeInvoices:
		return "tone-green"
	case categoryCodePhotos:
		return "tone-pink"
	default:
		return "tone-violet"
	}
}

func categoryLabelFromList(cats []dtos.DocumentCategoryResponse, code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	for _, c := range cats {
		if c.Code == code {
			return c.LabelES
		}
	}
	return code
}

func categoryInList(cats []dtos.DocumentCategoryResponse, code string) bool {
	code = strings.TrimSpace(code)
	for _, c := range cats {
		if c.Code == code {
			return true
		}
	}
	return false
}

func documentsBrowseURL(workspaceID, status string) string {
	q := url.Values{}
	if status != "" && status != entities.DocumentStatusActive {
		q.Set("status", status)
	}
	if workspaceID != "" {
		q.Set("workspace_id", workspaceID)
	}
	if len(q) == 0 {
		return "/archivo/documentos"
	}
	return "/archivo/documentos?" + q.Encode()
}

func documentsCreateURL(workspaceID, category string) string {
	q := url.Values{}
	if workspaceID != "" {
		q.Set("workspace_id", workspaceID)
	}
	if category != "" {
		q.Set("category_code", category)
	}
	if len(q) == 0 {
		return "/archivo/documentos/nuevo"
	}
	return "/archivo/documentos/nuevo?" + q.Encode()
}

func toFolderItem(f dtos.CategoryFolderResponse, workspaceID, status, view string) CategoryFolderItem {
	q := url.Values{}
	q.Set("category", f.Code)
	if status != "" {
		q.Set("status", status)
	}
	if view != "" {
		q.Set("view", view)
	}
	if workspaceID != "" {
		q.Set("workspace_id", workspaceID)
	}
	meta := "Sin documentos"
	if f.Count == 1 {
		meta = "1 documento"
	} else if f.Count > 1 {
		meta = fmt.Sprintf("%d documentos", f.Count)
	}
	return CategoryFolderItem{
		Code:     f.Code,
		Label:    f.LabelES,
		Count:    f.Count,
		OpenURL:  "/archivo/documentos?" + q.Encode(),
		IconTone: categoryIconTone(f.Code),
		MetaLine: meta,
	}
}

func documentMetaLine(item DocumentListItem) string {
	parts := make([]string, 0, documentMetaPartsCapacity)
	if item.CategoryLabel != "" {
		parts = append(parts, item.CategoryLabel)
	}
	if item.DocumentDate != "" {
		parts = append(parts, item.DocumentDate)
	}
	if item.AmountDisplay != "" {
		parts = append(parts, item.AmountDisplay)
	}
	if len(parts) == 0 {
		if item.Status == entities.DocumentStatusArchived {
			return "Archivado"
		}
		return "Sin fecha"
	}
	return strings.Join(parts, " · ")
}

func parseFormDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func optionalNonEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func formatAmountPesos(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	whole := cents / 100 //nolint:mnd
	frac := cents % 100  //nolint:mnd
	s := fmt.Sprintf("%d.%02d", whole, frac)
	if neg {
		return "-" + s
	}
	return s
}

func docsPaginationParser() *pagination.Parser {
	return pagination.NewParser(pagination.Config{
		DefaultLimit: defaultDocsPageLimit,
		MaxLimit:     100, //nolint:mnd
		MinLimit:     1,
	})
}

func archiveFlash(c echo.Context) string {
	if c.QueryParam("created") == "1" {
		return msgDocCreated
	}
	if c.QueryParam("updated") == "1" {
		return msgDocUpdated
	}
	return ""
}

func mapCreateWithFileError(err error) string {
	switch {
	case errors.Is(err, domainerrors.ErrFileRequired):
		return msgFileRequiredCreate
	case errors.Is(err, domainerrors.ErrFileTooLarge),
		errors.Is(err, domainerrors.ErrInvalidContentType),
		errors.Is(err, domainerrors.ErrTooManyFiles):
		return mapFileFormError(err)
	default:
		return mapDocFormError(err, msgDocCreateError)
	}
}

func mapDocFormError(err error, fallback string) string {
	switch {
	case errors.Is(err, domainerrors.ErrTitleRequired),
		errors.Is(err, domainerrors.ErrCategoryRequired),
		errors.Is(err, domainerrors.ErrInvalidCategory),
		errors.Is(err, domainerrors.ErrDocumentNotFound),
		errors.Is(err, domainerrors.ErrInsufficientWorkspaceRole):
		return err.Error()
	default:
		return fallback
	}
}

func mapHouseholdFormError(err error, fallback string) string {
	switch {
	case errors.Is(err, domainerrors.ErrWorkspaceNameRequired):
		return err.Error()
	default:
		return fallback
	}
}

func mapMemberFormError(err error, fallback string) string {
	switch {
	case errors.Is(err, domainerrors.ErrEmailRequired),
		errors.Is(err, domainerrors.ErrInvalidMemberRole),
		errors.Is(err, domainerrors.ErrCannotInviteSelf),
		errors.Is(err, domainerrors.ErrAlreadyMember),
		errors.Is(err, domainerrors.ErrCannotModifyOwner),
		errors.Is(err, domainerrors.ErrHouseholdOnlyInvite),
		errors.Is(err, domainerrors.ErrInviteeNotFound),
		errors.Is(err, domainerrors.ErrNotWorkspaceMember),
		errors.Is(err, domainerrors.ErrInsufficientWorkspaceRole):
		return err.Error()
	default:
		return fallback
	}
}

const (
	workspaceTypeLabelPersonal  = "Personal"
	workspaceTypeLabelHousehold = "Hogar"
)

func toWorkspaceListItem(ws *dtos.WorkspaceResponse) WorkspaceListItem {
	typeLabel := ws.Type
	switch ws.Type {
	case entities.WorkspaceTypePersonal:
		typeLabel = workspaceTypeLabelPersonal
	case entities.WorkspaceTypeHousehold:
		typeLabel = workspaceTypeLabelHousehold
	}
	item := WorkspaceListItem{
		ID:         ws.ID,
		Name:       ws.Name,
		Type:       ws.Type,
		TypeLabel:  typeLabel,
		MemberRole: ws.MemberRole,
		DocsURL:    "/archivo/documentos?workspace_id=" + ws.ID,
	}
	if ws.Type == entities.WorkspaceTypeHousehold && ws.MemberRole == entities.WorkspaceRoleOwner {
		item.MembersURL = "/archivo/hogares/" + ws.ID + "/miembros"
		item.ShowMembers = true
	}
	return item
}

func documentEditURL(docID, workspaceID string) string {
	url := "/archivo/documentos/" + docID + "/editar"
	if workspaceID != "" {
		url += "?workspace_id=" + workspaceID
	}
	return url
}

func (h *ArchivePageHandler) renderMembersWithError(c echo.Context, workspaceID string, inviteForm InviteMemberForm, errMsg string) error {
	userID := weblayout.CurrentUserID(c)
	layout := weblayout.AppLayoutFromEcho(c, archiveMembersTitle, "Administra quién puede ver y editar", "/archivo/hogares/"+workspaceID+"/miembros")

	workspaces, _ := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
	var ws *dtos.WorkspaceResponse
	for i := range workspaces {
		if workspaces[i].ID == workspaceID {
			ws = &workspaces[i]
			break
		}
	}
	members, _ := h.listMembersUC.Execute(c.Request().Context(), userID, workspaceID)

	return c.Render(http.StatusUnprocessableEntity, "archive/members", MembersPageData{
		AppLayoutData: layout,
		Workspace:     ws,
		Members:       members,
		CanManage:     ws != nil && ws.MemberRole == entities.WorkspaceRoleOwner,
		InviteForm:    inviteForm,
		Error:         errMsg,
	})
}
