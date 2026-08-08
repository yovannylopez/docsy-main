package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

// WorkspaceRepository persists workspaces and memberships.
type WorkspaceRepository interface {
	FindPersonalByOwner(ctx context.Context, ownerUserID string) (*entities.Workspace, error)
	FindByID(ctx context.Context, workspaceID string) (*entities.Workspace, error)
	ListForUser(ctx context.Context, userID string) ([]entities.WorkspaceMembership, error)
	CreateWorkspace(ctx context.Context, workspace *entities.Workspace) error
	AddMember(ctx context.Context, member *entities.WorkspaceMember) error
	FindMember(ctx context.Context, workspaceID, userID string) (*entities.WorkspaceMember, error)
	ListMembers(ctx context.Context, workspaceID string) ([]entities.WorkspaceMemberDetail, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
}

// EnsurePersonalWorkspaceService ensures the authenticated user has a personal workspace.
type EnsurePersonalWorkspaceService interface {
	Execute(ctx context.Context, userID string) (*dtos.WorkspaceResponse, error)
}

// ListWorkspacesService lists workspaces the user belongs to.
type ListWorkspacesService interface {
	Execute(ctx context.Context, userID string) ([]dtos.WorkspaceResponse, error)
}

// CreateHouseholdWorkspaceService creates a household workspace owned by the user.
type CreateHouseholdWorkspaceService interface {
	Execute(ctx context.Context, userID string, req *dtos.CreateHouseholdRequest) (*dtos.WorkspaceResponse, error)
}

// ListWorkspaceMembersService lists members of a workspace.
type ListWorkspaceMembersService interface {
	Execute(ctx context.Context, userID, workspaceID string) ([]dtos.WorkspaceMemberResponse, error)
}

// InviteWorkspaceMemberService adds an existing user to a household by email.
type InviteWorkspaceMemberService interface {
	Execute(ctx context.Context, actorUserID, workspaceID string, req *dtos.InviteMemberRequest) (*dtos.WorkspaceMemberResponse, error)
}

// UpdateWorkspaceMemberRoleService changes a member's role.
type UpdateWorkspaceMemberRoleService interface {
	Execute(ctx context.Context, actorUserID, workspaceID, targetUserID, role string) (*dtos.WorkspaceMemberResponse, error)
}

// RemoveWorkspaceMemberService removes a non-owner member.
type RemoveWorkspaceMemberService interface {
	Execute(ctx context.Context, actorUserID, workspaceID, targetUserID string) error
}
