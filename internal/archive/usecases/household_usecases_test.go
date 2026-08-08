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
	authmocks "github.com/yovannylopez/docsy-main/internal/auth/mocks"
)

type mockUserDirectory struct {
	mock.Mock
}

func (m *mockUserDirectory) FindByEmail(ctx context.Context, email string) (*ports.UserRef, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.UserRef), args.Error(1)
}

func TestCreateHouseholdWorkspaceUseCase_RequiresName(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	uc := NewCreateHouseholdWorkspaceUseCase(repo, nil)

	_, err := uc.Execute(context.Background(), "user-1", &dtos.CreateHouseholdRequest{Name: "  "})
	require.ErrorIs(t, err, domainerrors.ErrWorkspaceNameRequired)
}

func TestCreateHouseholdWorkspaceUseCase_Creates(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	audit := authmocks.NewAuditRepository(t)
	uc := NewCreateHouseholdWorkspaceUseCase(repo, audit)

	repo.On("CreateWorkspace", mock.Anything, mock.MatchedBy(func(w *entities.Workspace) bool {
		return w.Type == entities.WorkspaceTypeHousehold && w.Name == "Casa López"
	})).Return(nil)
	repo.On("AddMember", mock.Anything, mock.MatchedBy(func(m *entities.WorkspaceMember) bool {
		return m.UserID == "user-1" && m.Role == entities.WorkspaceRoleOwner
	})).Return(nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveHouseholdCreated)

	got, err := uc.Execute(context.Background(), "user-1", &dtos.CreateHouseholdRequest{Name: "Casa López"})
	require.NoError(t, err)
	assert.Equal(t, "Casa López", got.Name)
	assert.Equal(t, entities.WorkspaceTypeHousehold, got.Type)
	assert.Equal(t, entities.WorkspaceRoleOwner, got.MemberRole)
	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestInviteWorkspaceMemberUseCase_RequiresEmail(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	users := &mockUserDirectory{}
	uc := NewInviteWorkspaceMemberUseCase(repo, users, nil)
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.HouseholdWorkspace()
	owner := &entities.WorkspaceMember{UserID: "user-1", Role: entities.WorkspaceRoleOwner, JoinedAt: ws.CreatedAt}
	repo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(owner, nil)
	repo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)

	_, err := uc.Execute(context.Background(), "user-1", ws.ID, &dtos.InviteMemberRequest{Email: " "})
	require.ErrorIs(t, err, domainerrors.ErrEmailRequired)
}

func TestInviteWorkspaceMemberUseCase_InviteeNotFound(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	users := &mockUserDirectory{}
	uc := NewInviteWorkspaceMemberUseCase(repo, users, nil)
	stubs := archivetest.NewArchiveStubs()
	ws := stubs.HouseholdWorkspace()
	owner := &entities.WorkspaceMember{UserID: "user-1", Role: entities.WorkspaceRoleOwner, JoinedAt: ws.CreatedAt}
	repo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(owner, nil)
	repo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	users.On("FindByEmail", mock.Anything, "missing@example.com").Return((*ports.UserRef)(nil), nil)

	_, err := uc.Execute(context.Background(), "user-1", ws.ID, &dtos.InviteMemberRequest{Email: "missing@example.com"})
	require.ErrorIs(t, err, domainerrors.ErrInviteeNotFound)
}

func TestInviteWorkspaceMemberUseCase_RejectsPersonalWorkspace(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	users := &mockUserDirectory{}
	uc := NewInviteWorkspaceMemberUseCase(repo, users, nil)
	ws := archivetest.NewArchiveStubs().PersonalWorkspace()
	owner := archivetest.NewArchiveStubs().OwnerMember()
	repo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(owner, nil)
	repo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)

	_, err := uc.Execute(context.Background(), "user-1", ws.ID, archivetest.NewArchiveStubs().InviteMemberRequest())
	require.ErrorIs(t, err, domainerrors.ErrHouseholdOnlyInvite)
}

func TestInviteWorkspaceMemberUseCase_AlreadyMember(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	users := &mockUserDirectory{}
	uc := NewInviteWorkspaceMemberUseCase(repo, users, nil)
	ws := archivetest.NewArchiveStubs().HouseholdWorkspace()
	owner := &entities.WorkspaceMember{UserID: "user-1", Role: entities.WorkspaceRoleOwner, JoinedAt: ws.CreatedAt}
	existing := &entities.WorkspaceMember{UserID: "user-2", Role: entities.WorkspaceRoleMember, JoinedAt: ws.CreatedAt}
	repo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(owner, nil)
	repo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	users.On("FindByEmail", mock.Anything, "user2@example.com").Return(&ports.UserRef{ID: "user-2", Email: "user2@example.com"}, nil)
	repo.On("FindMember", mock.Anything, ws.ID, "user-2").Return(existing, nil)

	_, err := uc.Execute(context.Background(), "user-1", ws.ID, &dtos.InviteMemberRequest{Email: "user2@example.com"})
	require.ErrorIs(t, err, domainerrors.ErrAlreadyMember)
}

func TestInviteWorkspaceMemberUseCase_Invites(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	users := &mockUserDirectory{}
	audit := authmocks.NewAuditRepository(t)
	uc := NewInviteWorkspaceMemberUseCase(repo, users, audit)
	ws := archivetest.NewArchiveStubs().HouseholdWorkspace()
	owner := &entities.WorkspaceMember{UserID: "user-1", Role: entities.WorkspaceRoleOwner, JoinedAt: ws.CreatedAt}
	repo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(owner, nil)
	repo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	users.On("FindByEmail", mock.Anything, "member@example.com").Return(&ports.UserRef{
		ID: "user-2", Email: "member@example.com", DisplayName: "Member",
	}, nil)
	repo.On("FindMember", mock.Anything, ws.ID, "user-2").Return((*entities.WorkspaceMember)(nil), nil)
	repo.On("AddMember", mock.Anything, mock.MatchedBy(func(m *entities.WorkspaceMember) bool {
		return m.UserID == "user-2" && m.Role == entities.WorkspaceRoleMember
	})).Return(nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveMemberInvited)

	got, err := uc.Execute(context.Background(), "user-1", ws.ID, archivetest.NewArchiveStubs().InviteMemberRequest())
	require.NoError(t, err)
	assert.Equal(t, "user-2", got.UserID)
	assert.Equal(t, entities.WorkspaceRoleMember, got.Role)
	audit.AssertExpectations(t)
}

func TestUpdateWorkspaceMemberRoleUseCase_Updates(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	audit := authmocks.NewAuditRepository(t)
	uc := NewUpdateWorkspaceMemberRoleUseCase(repo, audit)
	ws := archivetest.NewArchiveStubs().HouseholdWorkspace()
	owner := &entities.WorkspaceMember{UserID: "user-1", Role: entities.WorkspaceRoleOwner, JoinedAt: ws.CreatedAt}
	member := &entities.WorkspaceMember{UserID: "user-2", Role: entities.WorkspaceRoleMember, JoinedAt: ws.CreatedAt}
	repo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(owner, nil)
	repo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	repo.On("FindMember", mock.Anything, ws.ID, "user-2").Return(member, nil)
	repo.On("UpdateMemberRole", mock.Anything, ws.ID, "user-2", entities.WorkspaceRoleViewer).Return(nil)
	repo.On("ListMembers", mock.Anything, ws.ID).Return([]entities.WorkspaceMemberDetail{{
		WorkspaceMember: entities.WorkspaceMember{UserID: "user-2", Role: entities.WorkspaceRoleViewer, JoinedAt: ws.CreatedAt},
		Email:           "member@example.com",
		FirstName:       "Member",
	}}, nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveMemberRoleUpdated)

	got, err := uc.Execute(context.Background(), "user-1", ws.ID, "user-2", entities.WorkspaceRoleViewer)
	require.NoError(t, err)
	assert.Equal(t, entities.WorkspaceRoleViewer, got.Role)
	audit.AssertExpectations(t)
}

func TestRemoveWorkspaceMemberUseCase_Removes(t *testing.T) {
	repo := &mockWorkspaceRepo{}
	audit := authmocks.NewAuditRepository(t)
	uc := NewRemoveWorkspaceMemberUseCase(repo, audit)
	ws := archivetest.NewArchiveStubs().HouseholdWorkspace()
	owner := &entities.WorkspaceMember{UserID: "user-1", Role: entities.WorkspaceRoleOwner, JoinedAt: ws.CreatedAt}
	member := &entities.WorkspaceMember{UserID: "user-2", Role: entities.WorkspaceRoleMember, JoinedAt: ws.CreatedAt}
	repo.On("FindMember", mock.Anything, ws.ID, "user-1").Return(owner, nil)
	repo.On("FindByID", mock.Anything, ws.ID).Return(ws, nil)
	repo.On("FindMember", mock.Anything, ws.ID, "user-2").Return(member, nil)
	repo.On("RemoveMember", mock.Anything, ws.ID, "user-2").Return(nil)
	expectAuditAction(t, audit, authdomain.AuditActionArchiveMemberRemoved)

	err := uc.Execute(context.Background(), "user-1", ws.ID, "user-2")
	require.NoError(t, err)
	audit.AssertExpectations(t)
}
