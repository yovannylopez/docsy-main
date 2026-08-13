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
	archivetest "github.com/yovannylopez/docsy-main/internal/archive/test_utils"
	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	authmocks "github.com/yovannylopez/docsy-main/internal/auth/mocks"
)

type mockDocumentRepo struct {
	mock.Mock
}

func (m *mockDocumentRepo) List(ctx context.Context, workspaceID string, filter dtos.ListDocumentsFilter) ([]entities.Document, int, error) {
	args := m.Called(ctx, workspaceID, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entities.Document), args.Int(1), args.Error(2)
}

func (m *mockDocumentRepo) FindByID(ctx context.Context, workspaceID, documentID string) (*entities.Document, error) {
	args := m.Called(ctx, workspaceID, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Document), args.Error(1)
}

func (m *mockDocumentRepo) Create(ctx context.Context, doc *entities.Document) error {
	return m.Called(ctx, doc).Error(0)
}

func (m *mockDocumentRepo) Update(ctx context.Context, doc *entities.Document) error {
	return m.Called(ctx, doc).Error(0)
}

func (m *mockDocumentRepo) Delete(ctx context.Context, workspaceID, documentID string) error {
	return m.Called(ctx, workspaceID, documentID).Error(0)
}

func (m *mockDocumentRepo) CategoryExists(ctx context.Context, workspaceID, code string) (bool, error) {
	args := m.Called(ctx, workspaceID, code)
	return args.Bool(0), args.Error(1)
}

func (m *mockDocumentRepo) ListCategories(ctx context.Context, workspaceID string) ([]entities.DocumentCategory, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DocumentCategory), args.Error(1)
}

func (m *mockDocumentRepo) FindCategory(ctx context.Context, workspaceID, code string) (*entities.DocumentCategory, error) {
	args := m.Called(ctx, workspaceID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentCategory), args.Error(1)
}

func (m *mockDocumentRepo) CreateCategory(ctx context.Context, cat *entities.DocumentCategory) error {
	return m.Called(ctx, cat).Error(0)
}

func (m *mockDocumentRepo) UpdateCategory(ctx context.Context, cat *entities.DocumentCategory) error {
	return m.Called(ctx, cat).Error(0)
}

func (m *mockDocumentRepo) UpdateSystemCategory(ctx context.Context, cat *entities.DocumentCategory) error {
	return m.Called(ctx, cat).Error(0)
}

func (m *mockDocumentRepo) DeactivateCategory(ctx context.Context, workspaceID, code string) error {
	return m.Called(ctx, workspaceID, code).Error(0)
}

func (m *mockDocumentRepo) CountCustomCategories(ctx context.Context, workspaceID string) (int, error) {
	args := m.Called(ctx, workspaceID)
	return args.Int(0), args.Error(1)
}

func (m *mockDocumentRepo) CountByCategory(ctx context.Context, workspaceID string, status string) (map[string]int, error) {
	args := m.Called(ctx, workspaceID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *mockDocumentRepo) CountDueAlerts(ctx context.Context, workspaceID, status string) (int, int, error) {
	args := m.Called(ctx, workspaceID, status)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockDocumentRepo) CountDueAlertsByCategory(ctx context.Context, workspaceID, status string) (map[string]dtos.CategoryDueAlertCounts, error) {
	args := m.Called(ctx, workspaceID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]dtos.CategoryDueAlertCounts), args.Error(1)
}

func stubEnsurePersonal(wsRepo *mockWorkspaceRepo) *EnsurePersonalWorkspaceUseCase {
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.PersonalWorkspace()
	member := stubs.OwnerMember()
	wsRepo.On("FindPersonalByOwner", mock.Anything, "user-1").Return(ws, nil)
	wsRepo.On("FindMember", mock.Anything, "ws-1", "user-1").Return(member, nil)
	return NewEnsurePersonalWorkspaceUseCase(wsRepo)
}

func expectAuditAction(t *testing.T, audit *authmocks.AuditRepository, action string) {
	t.Helper()
	audit.On("LogAction", mock.Anything, mock.MatchedBy(func(log *authentities.AuditLog) bool {
		return log.Action == action && log.Result == authdomain.AuditResultSuccess
	})).Return(nil).Once()
}

func TestCreateDocumentUseCase_RequiresTitle(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	ensure := stubEnsurePersonal(wsRepo)
	req := archivetest.NewArchiveStubs().CreateDocumentRequest()
	req.Title = "  "
	uc := NewCreateDocumentUseCase(wsRepo, ensure, docRepo, nil)

	_, err := uc.Execute(context.Background(), "user-1", req)
	require.ErrorIs(t, err, domainerrors.ErrTitleRequired)
}

func TestCreateDocumentUseCase_Creates(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	audit := authmocks.NewAuditRepository(t)
	ensure := stubEnsurePersonal(wsRepo)
	stubs := archivetest.NewArchiveStubs()
	uc := NewCreateDocumentUseCase(wsRepo, ensure, docRepo, audit)

	docRepo.On("CategoryExists", mock.Anything, mock.Anything, "taxes").Return(true, nil)
	docRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *entities.Document) bool {
		return d.Title == "Predial" && d.CategoryCode == "taxes" && d.WorkspaceID == "ws-1"
	})).Return(nil)
	docRepo.On("ListCategories", mock.Anything, mock.Anything).Return([]entities.DocumentCategory{
		{Code: "taxes", LabelES: "Impuestos"},
	}, nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveDocumentCreated)

	got, err := uc.Execute(context.Background(), "user-1", stubs.CreateDocumentRequest())
	require.NoError(t, err)
	assert.Equal(t, "Predial", got.Title)
	assert.Equal(t, "Impuestos", got.CategoryLabel)
	docRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestListDocumentsUseCase_Lists(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	ensure := stubEnsurePersonal(wsRepo)
	fileRepo := &mockFileRepo{}
	uc := NewListDocumentsUseCase(wsRepo, ensure, docRepo, fileRepo)
	doc := archivetest.CloneDocument(archivetest.NewArchiveStubs().ActiveDocument())
	doc.CategoryCode = "utilities"
	doc.Title = "Gas"

	docRepo.On("List", mock.Anything, "ws-1", mock.Anything).Return([]entities.Document{*doc}, 1, nil)
	docRepo.On("ListCategories", mock.Anything, mock.Anything).Return([]entities.DocumentCategory{
		{Code: "utilities", LabelES: "Servicios públicos"},
	}, nil)
	fileRepo.On("FindPrimaryByDocumentIDs", mock.Anything, []string{doc.ID}).Return(map[string]ports.DocumentPrimaryFile{
		doc.ID: {DocumentID: doc.ID, OriginalName: "gas.pdf", ContentType: mimeApplicationPDF},
	}, nil)

	docs, total, err := uc.Execute(context.Background(), "user-1", dtos.ListDocumentsFilter{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, docs, 1)
	assert.Equal(t, "Servicios públicos", docs[0].CategoryLabel)
	assert.Equal(t, "gas.pdf", docs[0].PrimaryOriginalName)
	assert.Equal(t, mimeApplicationPDF, docs[0].PrimaryContentType)
}

func TestListCategoryFoldersUseCase_ListsWithCounts(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	ensure := stubEnsurePersonal(wsRepo)
	uc := NewListCategoryFoldersUseCase(wsRepo, ensure, docRepo)

	docRepo.On("ListCategories", mock.Anything, mock.Anything).Return([]entities.DocumentCategory{
		{Code: "utilities", LabelES: "Servicios públicos", SortOrder: 1},
		{Code: "taxes", LabelES: "Impuestos", SortOrder: 2},
	}, nil)
	docRepo.On("CountByCategory", mock.Anything, "ws-1", "active").Return(map[string]int{
		"utilities": 3,
	}, nil)
	docRepo.On("CountDueAlertsByCategory", mock.Anything, "ws-1", "active").Return(map[string]dtos.CategoryDueAlertCounts{
		"utilities": {Upcoming: 1, Expired: 1},
	}, nil)

	folders, err := uc.Execute(context.Background(), "user-1", "", "active")
	require.NoError(t, err)
	require.Len(t, folders, 2)
	assert.Equal(t, 3, folders[0].Count)
	assert.Equal(t, 2, folders[0].AlertCount)
	assert.Equal(t, 1, folders[0].DueExpired)
	assert.Equal(t, 0, folders[1].Count)
	assert.Equal(t, "Servicios públicos", folders[0].LabelES)
}

func TestUpdateDocumentUseCase_Updates(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	audit := authmocks.NewAuditRepository(t)
	ensure := stubEnsurePersonal(wsRepo)
	uc := NewUpdateDocumentUseCase(wsRepo, ensure, docRepo, audit)
	doc := archivetest.CloneDocument(archivetest.NewArchiveStubs().ActiveDocument())

	docRepo.On("FindByID", mock.Anything, "ws-1", "d-1").Return(doc, nil)
	docRepo.On("Update", mock.Anything, mock.MatchedBy(func(d *entities.Document) bool {
		return d.Title == "Predial 2026"
	})).Return(nil)
	docRepo.On("ListCategories", mock.Anything, mock.Anything).Return([]entities.DocumentCategory{
		{Code: "taxes", LabelES: "Impuestos"},
	}, nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveDocumentUpdated)

	title := "Predial 2026"
	got, err := uc.Execute(context.Background(), "user-1", "", "d-1", &dtos.UpdateDocumentRequest{Title: &title})
	require.NoError(t, err)
	assert.Equal(t, "Predial 2026", got.Title)
	audit.AssertExpectations(t)
}

func TestArchiveDocumentUseCase_Archives(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	audit := authmocks.NewAuditRepository(t)
	ensure := stubEnsurePersonal(wsRepo)
	uc := NewArchiveDocumentUseCase(wsRepo, ensure, docRepo, audit)
	doc := archivetest.CloneDocument(archivetest.NewArchiveStubs().ActiveDocument())

	docRepo.On("FindByID", mock.Anything, "ws-1", "d-1").Return(doc, nil)
	docRepo.On("Update", mock.Anything, mock.MatchedBy(func(d *entities.Document) bool {
		return d.Status == entities.DocumentStatusArchived
	})).Return(nil)
	docRepo.On("ListCategories", mock.Anything, mock.Anything).Return([]entities.DocumentCategory{}, nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveDocumentArchived)

	got, err := uc.Execute(context.Background(), "user-1", "", "d-1")
	require.NoError(t, err)
	assert.Equal(t, entities.DocumentStatusArchived, got.Status)
	audit.AssertExpectations(t)
}

func TestGetDocumentUseCase_NotFound(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{}
	docRepo := &mockDocumentRepo{}
	ensure := stubEnsurePersonal(wsRepo)
	uc := NewGetDocumentUseCase(wsRepo, ensure, docRepo)

	docRepo.On("FindByID", mock.Anything, "ws-1", "missing").Return((*entities.Document)(nil), nil)

	_, err := uc.Execute(context.Background(), "user-1", "", "missing")
	require.ErrorIs(t, err, domainerrors.ErrDocumentNotFound)
}
