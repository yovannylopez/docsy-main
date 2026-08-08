package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// ListDocumentFiles handles GET /api/v1/archive/documents/:id/files
func (h *ArchiveHandler) ListDocumentFiles(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	files, err := h.listFilesUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), c.Param("id"))
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, files, "files retrieved successfully")
}

// UploadDocumentFile handles POST /api/v1/archive/documents/:id/files (multipart field "file")
func (h *ArchiveHandler) UploadDocumentFile(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}

	header, err := c.FormFile("file")
	if err != nil {
		return responses.BadRequest(c, domainerrors.ErrFileRequired.Error())
	}
	src, err := header.Open()
	if err != nil {
		return responses.BadRequest(c, domainerrors.ErrFileRequired.Error())
	}
	defer func() { _ = src.Close() }()

	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	file, err := h.uploadFileUC.Execute(
		c.Request().Context(),
		userID,
		workspaceIDParam(c),
		c.Param("id"),
		header.Filename,
		contentType,
		data,
	)
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.Created(c, file, "file uploaded successfully")
}

// DownloadDocumentFile handles GET /api/v1/archive/documents/:id/files/:fileId
func (h *ArchiveHandler) DownloadDocumentFile(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	result, err := h.downloadFileUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), c.Param("id"), c.Param("fileId"))
	if err != nil {
		return mapArchiveAPIError(c, err)
	}
	c.Response().Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, result.File.OriginalName),
	)
	c.Response().Header().Set("Content-Length", strconv.FormatInt(int64(len(result.Data)), 10))
	return c.Blob(http.StatusOK, result.File.ContentType, result.Data)
}

// DeleteDocumentFile handles DELETE /api/v1/archive/documents/:id/files/:fileId
func (h *ArchiveHandler) DeleteDocumentFile(c echo.Context) error {
	userID := currentUserID(c)
	if userID == "" {
		return responses.Unauthorized(c, "User not authenticated")
	}
	if err := h.deleteFileUC.Execute(c.Request().Context(), userID, workspaceIDParam(c), c.Param("id"), c.Param("fileId")); err != nil {
		return mapArchiveAPIError(c, err)
	}
	return responses.OK(c, nil, "file deleted successfully")
}
