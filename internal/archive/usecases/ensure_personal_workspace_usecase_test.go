package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	archivetest "github.com/yovannylopez/docsy-main/internal/archive/test_utils"
)

type mockWorkspaceRepo struct {
	mock.Mock
}

func (m *mockWorkspaceRepo) FindPersonalByOwner(ctx context.Context, ownerUserID string) (*entities.Workspace, error) {
	args := m.Called(ctx, ownerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Workspace), args.Error(1)
}

func (m *mockWorkspaceRepo) CreateWorkspace(ctx context.Context, workspace *entities.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *mockWorkspaceRepo) AddMember(ctx context.Context, member *entities.WorkspaceMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockWorkspaceRepo) FindMember(ctx context.Context, workspaceID, userID string) (*entities.WorkspaceMember, error) {
	args := m.Called(ctx, workspaceID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.WorkspaceMember), args.Error(1)
}

func (m *mockWorkspaceRepo) FindByID(ctx context.Context, workspaceID string) (*entities.Workspace, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Workspace), args.Error(1)
}

func (m *mockWorkspaceRepo) ListForUser(ctx context.Context, userID string) ([]entities.WorkspaceMembership, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.WorkspaceMembership), args.Error(1)
}

func (m *mockWorkspaceRepo) ListMembers(ctx context.Context, workspaceID string) ([]entities.WorkspaceMemberDetail, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.WorkspaceMemberDetail), args.Error(1)
}

func (m *mockWorkspaceRepo) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	args := m.Called(ctx, workspaceID, userID, role)
	return args.Error(0)
}

func (m *mockWorkspaceRepo) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

func TestEnsurePersonalWorkspaceUseCase_RequiresUserID(t *testing.T) {
	uc := NewEnsurePersonalWorkspaceUseCase(&mockWorkspaceRepo{})
	_, err := uc.Execute(context.Background(), "  ")
	require.ErrorIs(t, err, domainerrors.ErrUserIDRequired)
}

func TestEnsurePersonalWorkspaceUseCase_ReturnsExisting(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	uc := NewEnsurePersonalWorkspaceUseCase(repo)
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.PersonalWorkspace()
	member := stubs.OwnerMember()

	repo.On("FindPersonalByOwner", mock.Anything, "user-1").Return(ws, nil)
	repo.On("FindMember", mock.Anything, "ws-1", "user-1").Return(member, nil)

	got, err := uc.Execute(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, "ws-1", got.ID)
	assert.Equal(t, entities.WorkspaceRoleOwner, got.MemberRole)
	repo.AssertExpectations(t)
}

func TestEnsurePersonalWorkspaceUseCase_CreatesWhenMissing(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	uc := NewEnsurePersonalWorkspaceUseCase(repo)

	repo.On("FindPersonalByOwner", mock.Anything, "user-2").Return((*entities.Workspace)(nil), nil)
	repo.On("CreateWorkspace", mock.Anything, mock.MatchedBy(func(w *entities.Workspace) bool {
		return w.OwnerUserID == "user-2" && w.Type == entities.WorkspaceTypePersonal && w.Name == "Mi archivo"
	})).Return(nil)
	repo.On("AddMember", mock.Anything, mock.MatchedBy(func(m *entities.WorkspaceMember) bool {
		return m.UserID == "user-2" && m.Role == entities.WorkspaceRoleOwner
	})).Return(nil)

	got, err := uc.Execute(context.Background(), "user-2")
	require.NoError(t, err)
	assert.Equal(t, "user-2", got.OwnerUserID)
	assert.Equal(t, entities.WorkspaceTypePersonal, got.Type)
	assert.Equal(t, entities.WorkspaceRoleOwner, got.MemberRole)
	repo.AssertExpectations(t)
}
