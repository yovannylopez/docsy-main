package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/archive/infrastructure/storage"
	archivetest "github.com/yovannylopez/docsy-main/internal/archive/test_utils"
	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authmocks "github.com/yovannylopez/docsy-main/internal/auth/mocks"
)

type mockFileRepo struct {
	mock.Mock
}

func (m *mockFileRepo) Create(ctx context.Context, file *entities.DocumentFile) error {
	return m.Called(ctx, file).Error(0)
}

func (m *mockFileRepo) ListByDocument(ctx context.Context, documentID string) ([]entities.DocumentFile, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DocumentFile), args.Error(1)
}

func (m *mockFileRepo) FindByID(ctx context.Context, documentID, fileID string) (*entities.DocumentFile, error) {
	args := m.Called(ctx, documentID, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentFile), args.Error(1)
}

func (m *mockFileRepo) Delete(ctx context.Context, documentID, fileID string) error {
	return m.Called(ctx, documentID, fileID).Error(0)
}

func (m *mockFileRepo) CountByDocument(ctx context.Context, documentID string) (int, error) {
	args := m.Called(ctx, documentID)
	return args.Int(0), args.Error(1)
}

func (m *mockFileRepo) SumSizeBytesForUser(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockFileRepo) FindPrimaryByDocumentIDs(ctx context.Context, documentIDs []string) (map[string]ports.DocumentPrimaryFile, error) {
	args := m.Called(ctx, documentIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]ports.DocumentPrimaryFile), args.Error(1)
}

func TestCreateDocumentWithFileUseCase_RequiresFile(t *testing.T) {
	createUC := NewCreateDocumentUseCase(&mockWorkspaceRepo{}, stubEnsurePersonal(&mockWorkspaceRepo{}), &mockDocumentRepo{}, nil)
	uploadUC := NewUploadDocumentFileUseCase(
		&mockWorkspaceRepo{}, stubEnsurePersonal(&mockWorkspaceRepo{}), &mockDocumentRepo{},
		&mockFileRepo{}, storage.NewNoopDocumentStorage(), 1024, nil,
	)
	uc := NewCreateDocumentWithFileUseCase(createUC, uploadUC, &mockDocumentRepo{})
	_, _, err := uc.Execute(context.Background(), "user-1", &dtos.CreateDocumentRequest{
		CategoryCode: "taxes", Title: "Predial",
	}, "a.pdf", mimeApplicationPDF, nil)
	require.ErrorIs(t, err, domainerrors.ErrFileRequired)
}

func TestCreateDocumentWithFileUseCase_CreatesWithPDF(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	fileRepo := &mockFileRepo{}
	audit := authmocks.NewAuditRepository(t)
	ensure := stubEnsurePersonal(wsRepo)
	store, err := storage.NewLocalDocumentStorage(t.TempDir())
	require.NoError(t, err)

	ws := archivetest.NewArchiveStubs().PersonalWorkspace()
	createdDoc := archivetest.CloneDocument(archivetest.NewArchiveStubs().ActiveDocument())
	createdDoc.ID = "d-new"
	createdDoc.Title = "Predial"
	createdDoc.CategoryCode = "taxes"

	docRepo.On("CategoryExists", mock.Anything, mock.Anything, "taxes").Return(true, nil)
	docRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Document")).Run(func(args mock.Arguments) {
		d := args.Get(1).(*entities.Document)
		d.ID = "d-new"
	}).Return(nil)
	docRepo.On("ListCategories", mock.Anything, mock.Anything).Return([]entities.DocumentCategory{
		{Code: "taxes", LabelES: "Impuestos"},
	}, nil)
	wsRepo.On("FindByID", mock.Anything, "ws-1").Return(ws, nil)
	docRepo.On("FindByID", mock.Anything, "ws-1", "d-new").Return(createdDoc, nil)
	fileRepo.On("CountByDocument", mock.Anything, "d-new").Return(0, nil)
	fileRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DocumentFile")).Return(nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveDocumentCreated)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveFileUploaded)

	createUC := NewCreateDocumentUseCase(wsRepo, ensure, docRepo, audit)
	uploadUC := NewUploadDocumentFileUseCase(wsRepo, ensure, docRepo, fileRepo, store, 10*1024*1024, audit)
	uc := NewCreateDocumentWithFileUseCase(createUC, uploadUC, docRepo)

	pdf := []byte("%PDF-1.4")
	doc, file, err := uc.Execute(context.Background(), "user-1", &dtos.CreateDocumentRequest{
		CategoryCode: "taxes", Title: "Predial",
	}, "predial.pdf", mimeApplicationPDF, pdf)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotNil(t, file)
	assert.Equal(t, "Predial", doc.Title)
	assert.Equal(t, "predial.pdf", file.OriginalName)
}

func TestDeleteDocumentFileUseCase_RejectsLastFile(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	fileRepo := &mockFileRepo{}
	ensure := stubEnsurePersonal(wsRepo)
	store, err := storage.NewLocalDocumentStorage(t.TempDir())
	require.NoError(t, err)

	doc := archivetest.CloneDocument(archivetest.NewArchiveStubs().ActiveDocument())
	meta := archivetest.NewArchiveStubs().DocumentFile()
	docRepo.On("FindByID", mock.Anything, "ws-1", "d-1").Return(doc, nil)
	fileRepo.On("FindByID", mock.Anything, "d-1", "f-1").Return(meta, nil)
	fileRepo.On("CountByDocument", mock.Anything, "d-1").Return(1, nil)

	uc := NewDeleteDocumentFileUseCase(wsRepo, ensure, docRepo, fileRepo, store, nil)
	err = uc.Execute(context.Background(), "user-1", "", "d-1", "f-1")
	require.ErrorIs(t, err, domainerrors.ErrCannotDeleteLastFile)
}

func TestUploadDocumentFileUseCase_UploadsPDF(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	fileRepo := &mockFileRepo{}
	audit := authmocks.NewAuditRepository(t)
	ensure := stubEnsurePersonal(wsRepo)
	store, err := storage.NewLocalDocumentStorage(t.TempDir())
	require.NoError(t, err)

	doc := archivetest.CloneDocument(archivetest.NewArchiveStubs().ActiveDocument())
	docRepo.On("FindByID", mock.Anything, "ws-1", "d-1").Return(doc, nil)
	fileRepo.On("CountByDocument", mock.Anything, "d-1").Return(0, nil)
	fileRepo.On("Create", mock.Anything, mock.MatchedBy(func(f *entities.DocumentFile) bool {
		return f.DocumentID == "d-1" && f.ContentType == "application/pdf" && f.OriginalName == "predial.pdf"
	})).Return(nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveFileUploaded)

	uc := NewUploadDocumentFileUseCase(wsRepo, ensure, docRepo, fileRepo, store, 10*1024*1024, audit)
	pdf := []byte("%PDF-1.4 fake content for detect")
	got, err := uc.Execute(context.Background(), "user-1", "", "d-1", "predial.pdf", "application/pdf", pdf)
	require.NoError(t, err)
	assert.Equal(t, "predial.pdf", got.OriginalName)
	assert.Equal(t, "application/pdf", got.ContentType)
	fileRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestUploadDocumentFileUseCase_RejectsEmpty(t *testing.T) {
	uc := NewUploadDocumentFileUseCase(
		&mockWorkspaceRepo{}, stubEnsurePersonal(&mockWorkspaceRepo{}), &mockDocumentRepo{},
		&mockFileRepo{}, storage.NewNoopDocumentStorage(), 1024, nil,
	)
	_, err := uc.Execute(context.Background(), "user-1", "", "d-1", "a.pdf", "application/pdf", nil)
	require.ErrorIs(t, err, domainerrors.ErrFileRequired)
}

func TestUploadDocumentFileUseCase_RejectsTooLarge(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	ensure := stubEnsurePersonal(wsRepo)
	uc := NewUploadDocumentFileUseCase(wsRepo, ensure, &mockDocumentRepo{}, &mockFileRepo{}, storage.NewNoopDocumentStorage(), 4, nil)
	_, err := uc.Execute(context.Background(), "user-1", "", "d-1", "a.pdf", "application/pdf", []byte("12345"))
	require.ErrorIs(t, err, domainerrors.ErrFileTooLarge)
}

func TestDeleteDocumentFileUseCase_Deletes(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	fileRepo := &mockFileRepo{}
	audit := authmocks.NewAuditRepository(t)
	ensure := stubEnsurePersonal(wsRepo)
	store, err := storage.NewLocalDocumentStorage(t.TempDir())
	require.NoError(t, err)

	doc := archivetest.CloneDocument(archivetest.NewArchiveStubs().ActiveDocument())
	meta := archivetest.NewArchiveStubs().DocumentFile()
	require.NoError(t, store.Put(context.Background(), meta.StorageKey, meta.ContentType, []byte("%PDF")))

	docRepo.On("FindByID", mock.Anything, "ws-1", "d-1").Return(doc, nil)
	fileRepo.On("FindByID", mock.Anything, "d-1", "f-1").Return(meta, nil)
	fileRepo.On("CountByDocument", mock.Anything, "d-1").Return(2, nil)
	fileRepo.On("Delete", mock.Anything, "d-1", "f-1").Return(nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveFileDeleted)

	uc := NewDeleteDocumentFileUseCase(wsRepo, ensure, docRepo, fileRepo, store, audit)
	err = uc.Execute(context.Background(), "user-1", "", "d-1", "f-1")
	require.NoError(t, err)
	audit.AssertExpectations(t)
}

func TestSanitizeOriginalName(t *testing.T) {
	assert.Equal(t, "factura_2026.pdf", sanitizeOriginalName("../../factura 2026.pdf"))
	assert.Equal(t, "archivo", sanitizeOriginalName(".."))
}

func TestNormalizeContentType_AllowsJPEG(t *testing.T) {
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	ct, err := normalizeContentType("image/jpeg", "foto.jpg", jpeg)
	require.NoError(t, err)
	assert.Equal(t, mimeImageJPEG, ct)
}

func TestNormalizeContentType_AllowsOfficeAndTIFF(t *testing.T) {
	xlsx := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}
	ct, err := normalizeContentType("application/octet-stream", "libro.xlsx", xlsx)
	require.NoError(t, err)
	assert.Equal(t, mimeOOXMLSheet, ct)

	docx := []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00}
	ct, err = normalizeContentType("", "nota.docx", docx)
	require.NoError(t, err)
	assert.Equal(t, mimeOOXMLWord, ct)

	xls := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	ct, err = normalizeContentType(mimeMSExcel, "legacy.xls", xls)
	require.NoError(t, err)
	assert.Equal(t, mimeMSExcel, ct)

	tiff := []byte{0x49, 0x49, 0x2A, 0x00, 0x00}
	ct, err = normalizeContentType(mimeImageTIFF, "scan.tiff", tiff)
	require.NoError(t, err)
	assert.Equal(t, mimeImageTIFF, ct)
}

func TestNormalizeContentType_RejectsExtensionMismatch(t *testing.T) {
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	_, err := normalizeContentType(mimeImageJPEG, "fake.docx", jpeg)
	require.ErrorIs(t, err, domainerrors.ErrInvalidContentType)
}
