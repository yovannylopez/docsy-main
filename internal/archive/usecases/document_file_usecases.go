package usecases

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
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

const (
	maxAttachmentsPerDocument = 10
	maxOriginalNameLen        = 200
	mimeApplicationPDF        = "application/pdf"
	mimeImageJPEG             = "image/jpeg"
	mimeImagePNG              = "image/png"
	mimeImageWebP             = "image/webp"
	mimeImageGIF              = "image/gif"
	mimeImageTIFF             = "image/tiff"
	mimeMSExcel               = "application/vnd.ms-excel"
	mimeOOXMLSheet            = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	mimeMSWord                = "application/msword"
	mimeOOXMLWord             = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	extPDF  = ".pdf"
	extJPG  = ".jpg"
	extJPEG = ".jpeg"
	extPNG  = ".png"
	extWebP = ".webp"
	extGIF  = ".gif"
	extTIF  = ".tif"
	extTIFF = ".tiff"
	extXLS  = ".xls"
	extXLSX = ".xlsx"
	extDOC  = ".doc"
	extDOCX = ".docx"
)

func isAllowedContentType(ct string) bool {
	switch ct {
	case mimeApplicationPDF,
		mimeImageJPEG, mimeImagePNG, mimeImageWebP, mimeImageGIF, mimeImageTIFF,
		mimeMSExcel, mimeOOXMLSheet, mimeMSWord, mimeOOXMLWord:
		return true
	default:
		return false
	}
}

func isAllowedExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case extPDF,
		extJPG, extJPEG, extPNG, extWebP, extGIF, extTIF, extTIFF,
		extXLS, extXLSX,
		extDOC, extDOCX:
		return true
	default:
		return false
	}
}

func mimeFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case extPDF:
		return mimeApplicationPDF
	case extJPG, extJPEG:
		return mimeImageJPEG
	case extPNG:
		return mimeImagePNG
	case extWebP:
		return mimeImageWebP
	case extGIF:
		return mimeImageGIF
	case extTIF, extTIFF:
		return mimeImageTIFF
	case extXLS:
		return mimeMSExcel
	case extXLSX:
		return mimeOOXMLSheet
	case extDOC:
		return mimeMSWord
	case extDOCX:
		return mimeOOXMLWord
	default:
		return ""
	}
}

func hasPrefixBytes(data, prefix []byte) bool {
	return len(data) >= len(prefix) && bytes.Equal(data[:len(prefix)], prefix)
}

func matchesExtensionMagic(ext string, data []byte) bool {
	switch strings.ToLower(ext) {
	case extPDF:
		return hasPrefixBytes(data, []byte("%PDF"))
	case extJPG, extJPEG:
		return hasPrefixBytes(data, []byte{0xFF, 0xD8, 0xFF})
	case extPNG:
		return hasPrefixBytes(data, []byte{0x89, 0x50, 0x4E, 0x47})
	case extGIF:
		return hasPrefixBytes(data, []byte("GIF87a")) || hasPrefixBytes(data, []byte("GIF89a"))
	case extWebP:
		return len(data) >= 12 && hasPrefixBytes(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP"))
	case extTIF, extTIFF:
		return hasPrefixBytes(data, []byte{0x49, 0x49, 0x2A, 0x00}) || hasPrefixBytes(data, []byte{0x4D, 0x4D, 0x00, 0x2A})
	case extDOCX, extXLSX:
		// OOXML is a ZIP package.
		return hasPrefixBytes(data, []byte{0x50, 0x4B, 0x03, 0x04}) ||
			hasPrefixBytes(data, []byte{0x50, 0x4B, 0x05, 0x06}) ||
			hasPrefixBytes(data, []byte{0x50, 0x4B, 0x07, 0x08})
	case extDOC, extXLS:
		// Legacy OLE Compound File.
		return hasPrefixBytes(data, []byte{0xD0, 0xCF, 0x11, 0xE0})
	default:
		return false
	}
}

type documentFileAccess struct {
	documentAccess
	fileRepo ports.DocumentFileRepository
	storage  ports.DocumentStorage
	maxBytes int64
}

func toFileResponse(f *entities.DocumentFile) *dtos.DocumentFileResponse {
	return &dtos.DocumentFileResponse{
		ID:           f.ID,
		DocumentID:   f.DocumentID,
		OriginalName: f.OriginalName,
		ContentType:  f.ContentType,
		SizeBytes:    f.SizeBytes,
		UploadedAt:   f.UploadedAt,
	}
}

func (a *documentFileAccess) documentInWorkspace(
	ctx context.Context,
	userID, workspaceID, documentID string,
	needWrite bool,
) (*dtos.WorkspaceResponse, *entities.Document, error) {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return nil, nil, domainerrors.ErrDocumentIDRequired
	}
	ws, err := a.workspaceForUser(ctx, userID, workspaceID, needWrite)
	if err != nil {
		return nil, nil, err
	}
	doc, err := a.docRepo.FindByID(ctx, ws.ID, documentID)
	if err != nil {
		return nil, nil, err
	}
	if doc == nil {
		return nil, nil, domainerrors.ErrDocumentNotFound
	}
	return ws, doc, nil
}

// CreateDocumentWithFileUseCase creates a document together with its first required attachment.
type CreateDocumentWithFileUseCase struct {
	createUC *CreateDocumentUseCase
	uploadUC *UploadDocumentFileUseCase
	docRepo  ports.DocumentRepository
}

// NewCreateDocumentWithFileUseCase creates the use case.
func NewCreateDocumentWithFileUseCase(
	createUC *CreateDocumentUseCase,
	uploadUC *UploadDocumentFileUseCase,
	docRepo ports.DocumentRepository,
) *CreateDocumentWithFileUseCase {
	return &CreateDocumentWithFileUseCase{
		createUC: createUC,
		uploadUC: uploadUC,
		docRepo:  docRepo,
	}
}

// Execute creates the document and uploads the attachment; rolls back the document if upload fails.
func (uc *CreateDocumentWithFileUseCase) Execute(
	ctx context.Context,
	userID string,
	req *dtos.CreateDocumentRequest,
	originalName, contentType string,
	data []byte,
) (*dtos.DocumentResponse, *dtos.DocumentFileResponse, error) {
	if len(data) == 0 {
		return nil, nil, domainerrors.ErrFileRequired
	}
	if _, err := normalizeContentType(contentType, sanitizeOriginalName(originalName), data); err != nil {
		return nil, nil, err
	}

	doc, err := uc.createUC.Execute(ctx, userID, req)
	if err != nil {
		return nil, nil, err
	}

	file, err := uc.uploadUC.Execute(ctx, userID, doc.WorkspaceID, doc.ID, originalName, contentType, data)
	if err != nil {
		if delErr := uc.docRepo.Delete(ctx, doc.WorkspaceID, doc.ID); delErr != nil {
			return nil, nil, fmt.Errorf("upload attachment: %w (cleanup document: %v)", err, delErr)
		}
		return nil, nil, err
	}
	return doc, file, nil
}

// UploadDocumentFileUseCase stores a binary attachment for a document.
type UploadDocumentFileUseCase struct {
	access    documentFileAccess
	auditRepo authports.AuditRepository
}

// NewUploadDocumentFileUseCase creates the use case.
func NewUploadDocumentFileUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	fileRepo ports.DocumentFileRepository,
	storage ports.DocumentStorage,
	maxBytes int64,
	auditRepo authports.AuditRepository,
) *UploadDocumentFileUseCase {
	return &UploadDocumentFileUseCase{
		access: documentFileAccess{
			documentAccess: documentAccess{workspaceRepo, ensureUC, docRepo},
			fileRepo:       fileRepo,
			storage:        storage,
			maxBytes:       maxBytes,
		},
		auditRepo: auditRepo,
	}
}

// Execute validates and uploads a file.
func (uc *UploadDocumentFileUseCase) Execute(
	ctx context.Context,
	userID, workspaceID, documentID, originalName, contentType string,
	data []byte,
) (*dtos.DocumentFileResponse, error) {
	if len(data) == 0 {
		return nil, domainerrors.ErrFileRequired
	}
	if uc.access.maxBytes > 0 && int64(len(data)) > uc.access.maxBytes {
		return nil, domainerrors.ErrFileTooLarge
	}

	ws, doc, err := uc.access.documentInWorkspace(ctx, userID, workspaceID, documentID, true)
	if err != nil {
		return nil, err
	}

	count, err := uc.access.fileRepo.CountByDocument(ctx, doc.ID)
	if err != nil {
		return nil, err
	}
	if count >= maxAttachmentsPerDocument {
		return nil, domainerrors.ErrTooManyFiles
	}

	safeName := sanitizeOriginalName(originalName)
	normalizedType, err := normalizeContentType(contentType, safeName, data)
	if err != nil {
		return nil, err
	}
	fileID := uuid.NewString()
	key := fmt.Sprintf("%s/%s/%s_%s", ws.ID, doc.ID, fileID, safeName)

	if err := uc.access.storage.Put(ctx, key, normalizedType, data); err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}

	now := time.Now().UTC()
	uploadedBy := userID
	file := &entities.DocumentFile{
		ID:           fileID,
		DocumentID:   doc.ID,
		StorageKey:   key,
		OriginalName: safeName,
		ContentType:  normalizedType,
		SizeBytes:    int64(len(data)),
		UploadedBy:   &uploadedBy,
		UploadedAt:   now,
	}
	if err := uc.access.fileRepo.Create(ctx, file); err != nil {
		_ = uc.access.storage.Delete(ctx, key)
		return nil, fmt.Errorf("persist file metadata: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveFileUploaded,
		auditResourceDocumentFile, file.ID, "Archive document file uploaded successfully",
	)
	return toFileResponse(file), nil
}

// ListDocumentFilesUseCase lists attachments for a document.
type ListDocumentFilesUseCase struct {
	access documentFileAccess
}

// NewListDocumentFilesUseCase creates the use case.
func NewListDocumentFilesUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	fileRepo ports.DocumentFileRepository,
) *ListDocumentFilesUseCase {
	return &ListDocumentFilesUseCase{
		access: documentFileAccess{
			documentAccess: documentAccess{workspaceRepo, ensureUC, docRepo},
			fileRepo:       fileRepo,
		},
	}
}

// Execute returns attachment metadata.
func (uc *ListDocumentFilesUseCase) Execute(
	ctx context.Context,
	userID, workspaceID, documentID string,
) ([]dtos.DocumentFileResponse, error) {
	_, doc, err := uc.access.documentInWorkspace(ctx, userID, workspaceID, documentID, false)
	if err != nil {
		return nil, err
	}
	files, err := uc.access.fileRepo.ListByDocument(ctx, doc.ID)
	if err != nil {
		return nil, err
	}
	out := make([]dtos.DocumentFileResponse, 0, len(files))
	for i := range files {
		out = append(out, *toFileResponse(&files[i]))
	}
	return out, nil
}

// DownloadDocumentFileUseCase loads attachment bytes.
type DownloadDocumentFileUseCase struct {
	access documentFileAccess
}

// NewDownloadDocumentFileUseCase creates the use case.
func NewDownloadDocumentFileUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	fileRepo ports.DocumentFileRepository,
	storage ports.DocumentStorage,
) *DownloadDocumentFileUseCase {
	return &DownloadDocumentFileUseCase{
		access: documentFileAccess{
			documentAccess: documentAccess{workspaceRepo, ensureUC, docRepo},
			fileRepo:       fileRepo,
			storage:        storage,
		},
	}
}

// Execute returns file bytes and metadata.
func (uc *DownloadDocumentFileUseCase) Execute(
	ctx context.Context,
	userID, workspaceID, documentID, fileID string,
) (*dtos.DownloadDocumentFileResult, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, domainerrors.ErrFileIDRequired
	}
	_, doc, err := uc.access.documentInWorkspace(ctx, userID, workspaceID, documentID, false)
	if err != nil {
		return nil, err
	}
	meta, err := uc.access.fileRepo.FindByID(ctx, doc.ID, fileID)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, domainerrors.ErrFileNotFound
	}
	data, _, err := uc.access.storage.Get(ctx, meta.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("read stored file: %w", err)
	}
	return &dtos.DownloadDocumentFileResult{
		File: toFileResponse(meta),
		Data: data,
	}, nil
}

// DeleteDocumentFileUseCase removes an attachment.
type DeleteDocumentFileUseCase struct {
	access    documentFileAccess
	auditRepo authports.AuditRepository
}

// NewDeleteDocumentFileUseCase creates the use case.
func NewDeleteDocumentFileUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
	docRepo ports.DocumentRepository,
	fileRepo ports.DocumentFileRepository,
	storage ports.DocumentStorage,
	auditRepo authports.AuditRepository,
) *DeleteDocumentFileUseCase {
	return &DeleteDocumentFileUseCase{
		access: documentFileAccess{
			documentAccess: documentAccess{workspaceRepo, ensureUC, docRepo},
			fileRepo:       fileRepo,
			storage:        storage,
		},
		auditRepo: auditRepo,
	}
}

// Execute deletes metadata and the stored binary.
func (uc *DeleteDocumentFileUseCase) Execute(ctx context.Context, userID, workspaceID, documentID, fileID string) error {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return domainerrors.ErrFileIDRequired
	}
	_, doc, err := uc.access.documentInWorkspace(ctx, userID, workspaceID, documentID, true)
	if err != nil {
		return err
	}
	meta, err := uc.access.fileRepo.FindByID(ctx, doc.ID, fileID)
	if err != nil {
		return err
	}
	if meta == nil {
		return domainerrors.ErrFileNotFound
	}
	count, err := uc.access.fileRepo.CountByDocument(ctx, doc.ID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return domainerrors.ErrCannotDeleteLastFile
	}
	if err := uc.access.fileRepo.Delete(ctx, doc.ID, fileID); err != nil {
		return err
	}
	if err := uc.access.storage.Delete(ctx, meta.StorageKey); err != nil {
		return fmt.Errorf("delete stored file: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveFileDeleted,
		auditResourceDocumentFile, fileID, "Archive document file deleted successfully",
	)
	return nil
}

func normalizeContentType(declared, filename string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !isAllowedExtension(ext) {
		return "", domainerrors.ErrInvalidContentType
	}
	if !matchesExtensionMagic(ext, data) {
		return "", domainerrors.ErrInvalidContentType
	}

	canonical := mimeFromExtension(ext)
	if canonical == "" {
		return "", domainerrors.ErrInvalidContentType
	}

	declared = strings.ToLower(strings.TrimSpace(declared))
	if i := strings.Index(declared, ";"); i >= 0 {
		declared = strings.TrimSpace(declared[:i])
	}
	if declared == "image/jpg" {
		declared = mimeImageJPEG
	}

	detected := http.DetectContentType(data)
	if i := strings.Index(detected, ";"); i >= 0 {
		detected = strings.TrimSpace(detected[:i])
	}

	// Prefer a declared MIME when it is allowed and matches the extension family.
	if isAllowedContentType(declared) && mimeFamilyCompatible(declared, canonical) {
		return declared, nil
	}
	// Detected MIME is often wrong for Office (zip/octet-stream); keep extension canonical.
	if isAllowedContentType(detected) && mimeFamilyCompatible(detected, canonical) {
		return detected, nil
	}
	return canonical, nil
}

func mimeFamilyCompatible(got, want string) bool {
	if got == want {
		return true
	}
	// OOXML packages are frequently sniffed as ZIP / octet-stream.
	if (want == mimeOOXMLSheet || want == mimeOOXMLWord) &&
		(got == "application/zip" || got == "application/octet-stream") {
		return true
	}
	return false
}

func sanitizeOriginalName(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	if name == "." || name == ".." || name == "" {
		name = "archivo"
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteByte('_')
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" || out == "." || out == ".." {
		out = "archivo"
	}
	if len(out) > maxOriginalNameLen {
		ext := filepath.Ext(out)
		base := strings.TrimSuffix(out, ext)
		keep := maxOriginalNameLen - len(ext)
		if keep < 1 {
			keep = maxOriginalNameLen
			ext = ""
		}
		if len(base) > keep {
			base = base[:keep]
		}
		out = base + ext
	}
	return out
}
