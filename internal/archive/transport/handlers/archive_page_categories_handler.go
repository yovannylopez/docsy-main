package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
)

const (
	archiveCategoriesTitle    = "Categorías"
	archiveCategoriesSubtitle = "Organiza tu archivo con categorías planas"
	msgCategoriesLoadError    = "No se pudieron cargar las categorías. Intenta de nuevo."
	msgCategoryCreated        = "Categoría creada correctamente."
	msgCategoryUpdated        = "Categoría actualizada correctamente."
	msgCategoryDeactivated    = "Categoría desactivada correctamente."
	msgCategoryCreateError    = "No se pudo crear la categoría."
	msgCategoryUpdateError    = "No se pudo actualizar la categoría."
	msgCategoryDeleteError    = "No se pudo desactivar la categoría."
)

// CategoriesPageData holds the flat category management view.
type CategoriesPageData struct {
	weblayout.AppLayoutData
	WorkspaceID      string
	Workspaces       []dtos.WorkspaceResponse
	CustomCategories []CategoryManageItem
	SystemCategories []CategoryManageItem
	CanManageSystem  bool
	Form             CategoryForm
	Error            string
	Success          string
	DocsURL          string
	CancelEditURL    string
}

// CategoryManageItem is a row in the categories manager.
type CategoryManageItem struct {
	Code      string
	LabelES   string
	IsSystem  bool
	EditURL   string
	DeleteURL string
	RenameURL string
}

// CategoryForm holds create/rename values.
type CategoryForm struct {
	LabelES      string
	Code         string
	GeneralError string
}

// ShowCategories renders GET /archivo/categorias.
func (h *ArchivePageHandler) ShowCategories(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	workspaceID := workspaceIDParam(c)
	workspaces, _ := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
	cats, err := h.listCatsUC.Execute(c.Request().Context(), userID, workspaceID)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "archive/categories", CategoriesPageData{
			AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveCategoriesTitle, archiveCategoriesSubtitle, "/archivo/categorias"),
			WorkspaceID:   workspaceID,
			Workspaces:    workspaces,
			Error:         msgCategoriesLoadError,
			DocsURL:       documentsBrowseURL(workspaceID, ""),
		})
	}
	editCode := strings.TrimSpace(c.QueryParam("edit"))
	form := CategoryForm{}
	if editCode != "" {
		form.Code = editCode
		for _, cat := range cats {
			if cat.Code == editCode {
				form.LabelES = cat.LabelES
				break
			}
		}
	}
	return c.Render(http.StatusOK, "archive/categories", buildCategoriesPage(c, workspaceID, workspaces, cats, form, "", categoryFlash(c)))
}

// SubmitCreateCategory handles POST /archivo/categorias.
func (h *ArchivePageHandler) SubmitCreateCategory(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	workspaceID := workspaceIDParam(c)
	form := CategoryForm{LabelES: strings.TrimSpace(c.FormValue("label_es"))}
	_, err := h.createCatUC.Execute(c.Request().Context(), userID, &dtos.CreateCategoryRequest{
		WorkspaceID: workspaceID,
		LabelES:     form.LabelES,
	})
	if err != nil {
		workspaces, _ := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
		cats, _ := h.listCatsUC.Execute(c.Request().Context(), userID, workspaceID)
		form.GeneralError = mapCategoryFormError(err, msgCategoryCreateError)
		return c.Render(http.StatusUnprocessableEntity, "archive/categories", buildCategoriesPage(
			c, workspaceID, workspaces, cats, form, form.GeneralError, "",
		))
	}
	redirect := "/archivo/categorias?created=1"
	if workspaceID != "" {
		redirect += "&workspace_id=" + url.QueryEscape(workspaceID)
	}
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirect)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirect)
}

// SubmitUpdateCategory handles POST /archivo/categorias/:code/editar.
func (h *ArchivePageHandler) SubmitUpdateCategory(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	workspaceID := workspaceIDParam(c)
	code := c.Param("code")
	form := CategoryForm{
		Code:    code,
		LabelES: strings.TrimSpace(c.FormValue("label_es")),
	}
	_, err := h.updateCatUC.Execute(c.Request().Context(), userID, workspaceID, code, &dtos.UpdateCategoryRequest{
		WorkspaceID: workspaceID,
		LabelES:     form.LabelES,
	}, weblayout.CurrentUserIsSuperAdmin(c))
	if err != nil {
		workspaces, _ := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
		cats, _ := h.listCatsUC.Execute(c.Request().Context(), userID, workspaceID)
		form.GeneralError = mapCategoryFormError(err, msgCategoryUpdateError)
		return c.Render(http.StatusUnprocessableEntity, "archive/categories", buildCategoriesPage(
			c, workspaceID, workspaces, cats, form, form.GeneralError, "",
		))
	}
	redirect := "/archivo/categorias?updated=1"
	if workspaceID != "" {
		redirect += "&workspace_id=" + url.QueryEscape(workspaceID)
	}
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirect)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirect)
}

// SubmitDeactivateCategory handles POST /archivo/categorias/:code/desactivar.
func (h *ArchivePageHandler) SubmitDeactivateCategory(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	workspaceID := workspaceIDParam(c)
	code := c.Param("code")
	if err := h.deactivateCatUC.Execute(c.Request().Context(), userID, workspaceID, code); err != nil {
		workspaces, _ := h.listWorkspacesUC.Execute(c.Request().Context(), userID)
		cats, _ := h.listCatsUC.Execute(c.Request().Context(), userID, workspaceID)
		msg := mapCategoryFormError(err, msgCategoryDeleteError)
		return c.Render(http.StatusUnprocessableEntity, "archive/categories", buildCategoriesPage(
			c, workspaceID, workspaces, cats, CategoryForm{}, msg, "",
		))
	}
	redirect := "/archivo/categorias?deleted=1"
	if workspaceID != "" {
		redirect += "&workspace_id=" + url.QueryEscape(workspaceID)
	}
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirect)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirect)
}

func buildCategoriesPage(
	c echo.Context,
	workspaceID string,
	workspaces []dtos.WorkspaceResponse,
	cats []dtos.DocumentCategoryResponse,
	form CategoryForm,
	errMsg, success string,
) CategoriesPageData {
	custom := make([]CategoryManageItem, 0, len(cats))
	system := make([]CategoryManageItem, 0, len(cats))
	canManageSystem := weblayout.CurrentUserIsSuperAdmin(c)
	baseQ := url.Values{}
	if workspaceID != "" {
		baseQ.Set("workspace_id", workspaceID)
	}
	listURL := "/archivo/categorias"
	if encoded := baseQ.Encode(); encoded != "" {
		listURL += "?" + encoded
	}

	for _, cat := range cats {
		item := CategoryManageItem{
			Code:     cat.Code,
			LabelES:  cat.LabelES,
			IsSystem: cat.IsSystem,
		}
		q := ""
		if workspaceID != "" {
			q = "?workspace_id=" + url.QueryEscape(workspaceID)
		}
		if !cat.IsSystem || canManageSystem {
			item.EditURL = fmt.Sprintf("/archivo/categorias/%s/editar%s", url.PathEscape(cat.Code), q)
			renameQ := url.Values{}
			if workspaceID != "" {
				renameQ.Set("workspace_id", workspaceID)
			}
			renameQ.Set("edit", cat.Code)
			item.RenameURL = "/archivo/categorias?" + renameQ.Encode()
		}
		if !cat.IsSystem {
			item.DeleteURL = fmt.Sprintf("/archivo/categorias/%s/desactivar%s", url.PathEscape(cat.Code), q)
			custom = append(custom, item)
			continue
		}
		system = append(system, item)
	}

	return CategoriesPageData{
		AppLayoutData:    weblayout.AppLayoutFromEcho(c, archiveCategoriesTitle, archiveCategoriesSubtitle, "/archivo/categorias"),
		WorkspaceID:      workspaceID,
		Workspaces:       workspaces,
		CustomCategories: custom,
		SystemCategories: system,
		CanManageSystem:  canManageSystem,
		Form:             form,
		Error:            errMsg,
		Success:          success,
		DocsURL:          documentsBrowseURL(workspaceID, ""),
		CancelEditURL:    listURL,
	}
}

func categoryFlash(c echo.Context) string {
	switch {
	case c.QueryParam("created") == "1":
		return msgCategoryCreated
	case c.QueryParam("updated") == "1":
		return msgCategoryUpdated
	case c.QueryParam("deleted") == "1":
		return msgCategoryDeactivated
	default:
		return ""
	}
}

func mapCategoryFormError(err error, fallback string) string {
	switch {
	case errors.Is(err, domainerrors.ErrCategoryLabelRequired):
		return "El nombre de la categoría es obligatorio."
	case errors.Is(err, domainerrors.ErrCategoryLabelTooLong):
		return "El nombre de la categoría es demasiado largo."
	case errors.Is(err, domainerrors.ErrCategoryDuplicateLabel):
		return "Ya existe una categoría con ese nombre."
	case errors.Is(err, domainerrors.ErrTooManyCustomCategories):
		return "Alcanzaste el máximo de categorías personalizadas (20)."
	case errors.Is(err, domainerrors.ErrCannotModifySystemCategory):
		return "Solo un super administrador puede editar categorías base."
	case errors.Is(err, domainerrors.ErrCategoryInUse):
		return "La categoría tiene documentos. Muévelos antes de desactivarla."
	case errors.Is(err, domainerrors.ErrCategoryNotFound):
		return "Categoría no encontrada."
	case errors.Is(err, domainerrors.ErrInsufficientWorkspaceRole),
		errors.Is(err, domainerrors.ErrNotWorkspaceMember):
		return "No tienes permiso para gestionar categorías en este archivo."
	default:
		return fallback
	}
}
