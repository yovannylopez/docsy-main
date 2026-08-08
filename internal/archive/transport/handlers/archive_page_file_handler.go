package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
)

// UploadDocumentFile handles POST /archivo/documentos/:id/archivos
func (h *ArchivePageHandler) UploadDocumentFile(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	docID := c.Param("id")
	redirectEdit := fmt.Sprintf("/archivo/documentos/%s/editar", docID)

	header, err := c.FormFile("file")
	if err != nil {
		return h.renderEditWithFileError(c, userID, docID, domainerrors.ErrFileRequired.Error())
	}
	src, err := header.Open()
	if err != nil {
		return h.renderEditWithFileError(c, userID, docID, msgFileUploadError)
	}
	defer func() { _ = src.Close() }()

	data, err := io.ReadAll(src)
	if err != nil {
		return h.renderEditWithFileError(c, userID, docID, msgFileUploadError)
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	_, err = h.uploadFileUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), docID, header.Filename, contentType, data)
	if err != nil {
		return h.renderEditWithFileError(c, userID, docID, mapFileFormError(err))
	}

	redirectURL := redirectEdit + "?uploaded=1"
	if wsID := workspaceIDParam(c); wsID != "" {
		redirectURL += "&workspace_id=" + wsID
	}
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

// DownloadDocumentFile handles GET /archivo/documentos/:id/archivos/:fileId
// With ?inline=1 serves Content-Disposition: inline for in-browser preview.
func (h *ArchivePageHandler) DownloadDocumentFile(c echo.Context) error {
	inline := strings.EqualFold(c.QueryParam("inline"), "1") || strings.EqualFold(c.QueryParam("inline"), "true")
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		if inline {
			return c.NoContent(http.StatusUnauthorized)
		}
		return c.Redirect(http.StatusFound, "/login")
	}
	result, err := h.downloadFileUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), c.Param("id"), c.Param("fileId"))
	if err != nil {
		if inline {
			return c.NoContent(http.StatusNotFound)
		}
		return c.Redirect(http.StatusFound, "/archivo/documentos/"+c.Param("id")+"/editar")
	}
	disposition := "attachment"
	if inline {
		disposition = "inline"
	}
	c.Response().Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`%s; filename="%s"`, disposition, result.File.OriginalName),
	)
	c.Response().Header().Set("Content-Length", strconv.FormatInt(int64(len(result.Data)), 10))
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("Cache-Control", "private, no-store")
	contentType := result.File.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(result.Data)
	}
	return c.Blob(http.StatusOK, contentType, result.Data)
}

// DeleteDocumentFile handles POST /archivo/documentos/:id/archivos/:fileId/eliminar
func (h *ArchivePageHandler) DeleteDocumentFile(c echo.Context) error {
	userID := weblayout.CurrentUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}
	docID := c.Param("id")
	if err := h.deleteFileUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), docID, c.Param("fileId")); err != nil {
		return h.renderEditWithFileError(c, userID, docID, mapFileFormError(err))
	}
	redirectURL := fmt.Sprintf("/archivo/documentos/%s/editar?deleted=1", docID)
	if wsID := workspaceIDParam(c); wsID != "" {
		redirectURL += "&workspace_id=" + wsID
	}
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

func (h *ArchivePageHandler) renderEditWithFileError(c echo.Context, userID, docID, errMsg string) error {
	wsID := workspaceIDParam(c)
	doc, err := h.getDocUC.Execute(c.Request().Context(), userID, wsID, docID)
	cats, _ := h.listCatsUC.Execute(c.Request().Context())
	if err != nil {
		return h.renderForm(c, DocumentForm{ID: docID, WorkspaceID: wsID}, cats, true, errMsg, http.StatusUnprocessableEntity)
	}
	form := documentToForm(doc)
	form.GeneralError = errMsg
	return c.Render(http.StatusUnprocessableEntity, "archive/document_form", DocumentFormPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, archiveEditTitle, "Actualiza los metadatos", "/archivo/documentos/"+docID+"/editar"),
		WorkspaceID:   doc.WorkspaceID,
		Form:          form,
		Categories:    cats,
		Files:         h.loadFileItems(c.Request().Context(), userID, doc.WorkspaceID, docID),
		Error:         errMsg,
		IsEdit:        true,
	})
}

func (h *ArchivePageHandler) loadFileItems(ctx context.Context, userID, workspaceID, docID string) []DocumentFileItem {
	if h.listFilesUC == nil {
		return nil
	}
	files, err := h.listFilesUC.Execute(ctx, userID, workspaceID, docID)
	if err != nil {
		return nil
	}
	items := make([]DocumentFileItem, 0, len(files))
	for _, f := range files {
		ext := fileExtension(f.OriginalName)
		kind := previewKind(f.ContentType, f.OriginalName)
		downloadURL := fmt.Sprintf("/archivo/documentos/%s/archivos/%s", docID, f.ID)
		previewURL := downloadURL + "?inline=1"
		if workspaceID != "" {
			downloadURL += "?workspace_id=" + workspaceID
			previewURL += "&workspace_id=" + workspaceID
		}
		items = append(items, DocumentFileItem{
			ID:            f.ID,
			OriginalName:  f.OriginalName,
			ContentType:   f.ContentType,
			SizeDisplay:   formatFileSize(f.SizeBytes),
			UploadedAt:    f.UploadedAt.Format("02/01/2006"),
			DownloadURL:   downloadURL,
			PreviewURL:    previewURL,
			DeleteURL:     fmt.Sprintf("/archivo/documentos/%s/archivos/%s/eliminar", docID, f.ID),
			Extension:     ext,
			BadgeColor:    fileExtensionBadgeColor(ext),
			PreviewKind:   kind,
			IsPreviewable: kind == previewKindPDF || kind == previewKindImage,
		})
	}
	return items
}

func fileExtension(filename string) string {
	ext := strings.TrimPrefix(strings.ToUpper(path.Ext(filename)), ".")
	if ext == "" {
		return "FILE"
	}
	return ext
}

func fileExtensionBadgeColor(ext string) string {
	switch strings.ToLower(ext) {
	case "pdf":
		return "#EF4444"
	case "jpg", "jpeg", "png", "gif", "webp", "tif", "tiff":
		return "#8B5CF6"
	case "xls", "xlsx":
		return "#059669"
	case "doc", "docx":
		return "#2563EB"
	default:
		return "#6366F1"
	}
}

const (
	previewKindPDF   = "pdf"
	previewKindImage = "image"
	previewKindOther = "other"
)

func previewKind(contentType, filename string) string {
	ct := strings.ToLower(contentType)
	name := strings.ToLower(filename)
	switch {
	case strings.Contains(ct, "pdf") || strings.HasSuffix(name, ".pdf"):
		return previewKindPDF
	case strings.Contains(ct, "tiff") || strings.HasSuffix(name, ".tif") || strings.HasSuffix(name, ".tiff"):
		// Browsers rarely render TIFF inline; treat as download-only.
		return previewKindOther
	case strings.HasPrefix(ct, "image/") ||
		strings.HasSuffix(name, ".jpg") ||
		strings.HasSuffix(name, ".jpeg") ||
		strings.HasSuffix(name, ".png") ||
		strings.HasSuffix(name, ".gif") ||
		strings.HasSuffix(name, ".webp"):
		return previewKindImage
	default:
		return previewKindOther
	}
}

func formatFileSize(n int64) string {
	const (
		kb = 1024
		mb = 1024 * 1024
	)
	switch {
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func mapFileFormError(err error) string {
	switch {
	case errors.Is(err, domainerrors.ErrFileRequired):
		return msgFileRequiredCreate
	case errors.Is(err, domainerrors.ErrCannotDeleteLastFile):
		return msgCannotDeleteLastFile
	case errors.Is(err, domainerrors.ErrFileTooLarge),
		errors.Is(err, domainerrors.ErrInvalidContentType),
		errors.Is(err, domainerrors.ErrTooManyFiles),
		errors.Is(err, domainerrors.ErrFileNotFound),
		errors.Is(err, domainerrors.ErrDocumentNotFound):
		return err.Error()
	default:
		return msgFileUploadError
	}
}
