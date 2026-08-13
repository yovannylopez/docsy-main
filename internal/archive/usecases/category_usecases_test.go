package usecases

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	archivetest "github.com/yovannylopez/docsy-main/internal/archive/test_utils"
)

func TestCreateCategoryUseCase_CreatesCustomFlatCategory(t *testing.T) {
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.PersonalWorkspace()
	wsRepo := &mockWorkspaceRepo{}
	wsRepo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	wsRepo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(stubs.OwnerMember(), nil)

	docRepo := &mockDocumentRepo{}
	docRepo.On("CountCustomCategories", mock.Anything, ws.ID).Return(0, nil)
	docRepo.On("ListCategories", mock.Anything, ws.ID).Return([]entities.DocumentCategory{
		{Code: "taxes", LabelES: "Impuestos", IsSystem: true, IsActive: true},
	}, nil)
	docRepo.On("CreateCategory", mock.Anything, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil)

	uc := NewCreateCategoryUseCase(wsRepo, nil, docRepo, nil)
	got, err := uc.Execute(context.Background(), "user-1", &dtos.CreateCategoryRequest{
		WorkspaceID: ws.ID,
		LabelES:     "Certificados escolares",
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Certificados escolares", got.LabelES)
	assert.False(t, got.IsSystem)
	assert.True(t, strings.HasPrefix(got.Code, "c_"))
}

func TestCreateCategoryUseCase_RejectsDuplicateLabel(t *testing.T) {
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.PersonalWorkspace()
	wsRepo := &mockWorkspaceRepo{}
	wsRepo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	wsRepo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(stubs.OwnerMember(), nil)

	docRepo := &mockDocumentRepo{}
	docRepo.On("CountCustomCategories", mock.Anything, ws.ID).Return(1, nil)
	docRepo.On("ListCategories", mock.Anything, ws.ID).Return([]entities.DocumentCategory{
		{Code: "taxes", LabelES: "Impuestos", IsSystem: true, IsActive: true},
	}, nil)

	uc := NewCreateCategoryUseCase(wsRepo, nil, docRepo, nil)
	_, err := uc.Execute(context.Background(), "user-1", &dtos.CreateCategoryRequest{
		WorkspaceID: ws.ID,
		LabelES:     "impuestos",
	})
	assert.ErrorIs(t, err, domainerrors.ErrCategoryDuplicateLabel)
}

func TestUpdateCategoryUseCase_SuperAdminRenamesSystem(t *testing.T) {
	docRepo := &mockDocumentRepo{}
	docRepo.On("ListCategories", mock.Anything, mock.Anything).Return([]entities.DocumentCategory{
		{Code: "taxes", LabelES: "Impuestos", IsSystem: true, IsActive: true},
		{Code: "health", LabelES: "Salud", IsSystem: true, IsActive: true},
	}, nil)
	docRepo.On("UpdateSystemCategory", mock.Anything, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil)

	uc := NewUpdateCategoryUseCase(&mockWorkspaceRepo{}, nil, docRepo, nil)
	got, err := uc.Execute(context.Background(), "admin-1", "", "taxes", &dtos.UpdateCategoryRequest{
		LabelES: "Impuestos prediales",
	}, true)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Impuestos prediales", got.LabelES)
	assert.True(t, got.IsSystem)
}

func TestUpdateCategoryUseCase_RejectsSystemWithoutAllow(t *testing.T) {
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.PersonalWorkspace()
	wsRepo := &mockWorkspaceRepo{}
	wsRepo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	wsRepo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(stubs.OwnerMember(), nil)

	docRepo := &mockDocumentRepo{}
	docRepo.On("FindCategory", mock.Anything, ws.ID, "taxes").Return(&entities.DocumentCategory{
		Code: "taxes", LabelES: "Impuestos", IsSystem: true, IsActive: true,
	}, nil)

	uc := NewUpdateCategoryUseCase(wsRepo, nil, docRepo, nil)
	_, err := uc.Execute(context.Background(), "user-1", ws.ID, "taxes", &dtos.UpdateCategoryRequest{
		LabelES: "Otro",
	}, false)
	assert.ErrorIs(t, err, domainerrors.ErrCannotModifySystemCategory)
}

func TestDeactivateCategoryUseCase_BlocksSystem(t *testing.T) {
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.PersonalWorkspace()
	wsRepo := &mockWorkspaceRepo{}
	wsRepo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	wsRepo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(stubs.OwnerMember(), nil)

	docRepo := &mockDocumentRepo{}
	docRepo.On("FindCategory", mock.Anything, ws.ID, "taxes").Return(&entities.DocumentCategory{
		Code: "taxes", LabelES: "Impuestos", IsSystem: true, IsActive: true,
	}, nil)

	uc := NewDeactivateCategoryUseCase(wsRepo, nil, docRepo, nil)
	err := uc.Execute(context.Background(), "user-1", ws.ID, "taxes")
	assert.ErrorIs(t, err, domainerrors.ErrCannotModifySystemCategory)
}
