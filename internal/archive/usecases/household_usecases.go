package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

const defaultHouseholdWorkspaceName = "Archivo del hogar"

// ListWorkspacesUseCase lists workspaces for a user (ensures personal exists first).
type ListWorkspacesUseCase struct {
	workspaceRepo ports.WorkspaceRepository
	ensureUC      ports.EnsurePersonalWorkspaceService
}

// NewListWorkspacesUseCase creates the use case.
func NewListWorkspacesUseCase(
	workspaceRepo ports.WorkspaceRepository,
	ensureUC ports.EnsurePersonalWorkspaceService,
) *ListWorkspacesUseCase {
	return &ListWorkspacesUseCase{workspaceRepo: workspaceRepo, ensureUC: ensureUC}
}

// Execute returns all workspaces the user belongs to.
func (uc *ListWorkspacesUseCase) Execute(ctx context.Context, userID string) ([]dtos.WorkspaceResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}
	if _, err := uc.ensureUC.Execute(ctx, userID); err != nil {
		return nil, err
	}
	rows, err := uc.workspaceRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dtos.WorkspaceResponse, 0, len(rows))
	for i := range rows {
		out = append(out, *toWorkspaceResponse(&rows[i].Workspace, rows[i].MemberRole))
	}
	return out, nil
}

// CreateHouseholdWorkspaceUseCase creates a household archive.
type CreateHouseholdWorkspaceUseCase struct {
	workspaceRepo ports.WorkspaceRepository
	auditRepo     authports.AuditRepository
}

// NewCreateHouseholdWorkspaceUseCase creates the use case.
func NewCreateHouseholdWorkspaceUseCase(
	workspaceRepo ports.WorkspaceRepository,
	auditRepo authports.AuditRepository,
) *CreateHouseholdWorkspaceUseCase {
	return &CreateHouseholdWorkspaceUseCase{workspaceRepo: workspaceRepo, auditRepo: auditRepo}
}

// Execute creates the household and owner membership.
func (uc *CreateHouseholdWorkspaceUseCase) Execute(
	ctx context.Context,
	userID string,
	req *dtos.CreateHouseholdRequest,
) (*dtos.WorkspaceResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}
	name := defaultHouseholdWorkspaceName
	if req != nil {
		name = strings.TrimSpace(req.Name)
	}
	if name == "" {
		return nil, domainerrors.ErrWorkspaceNameRequired
	}

	now := time.Now().UTC()
	ws := &entities.Workspace{
		ID:          uuid.NewString(),
		Name:        name,
		Type:        entities.WorkspaceTypeHousehold,
		OwnerUserID: userID,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := uc.workspaceRepo.CreateWorkspace(ctx, ws); err != nil {
		return nil, fmt.Errorf("create household workspace: %w", err)
	}
	member := &entities.WorkspaceMember{
		ID:          uuid.NewString(),
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        entities.WorkspaceRoleOwner,
		JoinedAt:    now,
	}
	if err := uc.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("add household owner: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, userID,
		authdomain.AuditActionArchiveHouseholdCreated,
		auditResourceWorkspace, ws.ID, "Archive household workspace created successfully",
	)
	return toWorkspaceResponse(ws, entities.WorkspaceRoleOwner), nil
}

// ListWorkspaceMembersUseCase lists members of a workspace the actor belongs to.
type ListWorkspaceMembersUseCase struct {
	workspaceRepo ports.WorkspaceRepository
}

// NewListWorkspaceMembersUseCase creates the use case.
func NewListWorkspaceMembersUseCase(workspaceRepo ports.WorkspaceRepository) *ListWorkspaceMembersUseCase {
	return &ListWorkspaceMembersUseCase{workspaceRepo: workspaceRepo}
}

// Execute lists members.
func (uc *ListWorkspaceMembersUseCase) Execute(
	ctx context.Context,
	actorUserID, workspaceID string,
) ([]dtos.WorkspaceMemberResponse, error) {
	if _, err := requireMembership(ctx, uc.workspaceRepo, actorUserID, workspaceID); err != nil {
		return nil, err
	}
	rows, err := uc.workspaceRepo.ListMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]dtos.WorkspaceMemberResponse, 0, len(rows))
	for i := range rows {
		out = append(out, toMemberResponse(&rows[i]))
	}
	return out, nil
}

// InviteWorkspaceMemberUseCase adds an existing user to a household.
type InviteWorkspaceMemberUseCase struct {
	workspaceRepo ports.WorkspaceRepository
	users         ports.UserDirectory
	auditRepo     authports.AuditRepository
}

// NewInviteWorkspaceMemberUseCase creates the use case.
func NewInviteWorkspaceMemberUseCase(
	workspaceRepo ports.WorkspaceRepository,
	users ports.UserDirectory,
	auditRepo authports.AuditRepository,
) *InviteWorkspaceMemberUseCase {
	return &InviteWorkspaceMemberUseCase{workspaceRepo: workspaceRepo, users: users, auditRepo: auditRepo}
}

// Execute invites by email.
func (uc *InviteWorkspaceMemberUseCase) Execute(
	ctx context.Context,
	actorUserID, workspaceID string,
	req *dtos.InviteMemberRequest,
) (*dtos.WorkspaceMemberResponse, error) {
	if err := requireOwner(ctx, uc.workspaceRepo, actorUserID, workspaceID); err != nil {
		return nil, err
	}

	ws, err := uc.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, domainerrors.ErrWorkspaceNotFound
	}
	if ws.Type != entities.WorkspaceTypeHousehold {
		return nil, domainerrors.ErrHouseholdOnlyInvite
	}

	if req == nil {
		return nil, domainerrors.ErrEmailRequired
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, domainerrors.ErrEmailRequired
	}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = entities.WorkspaceRoleMember
	}
	if role != entities.WorkspaceRoleMember && role != entities.WorkspaceRoleViewer {
		return nil, domainerrors.ErrInvalidMemberRole
	}

	invitee, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if invitee == nil {
		return nil, domainerrors.ErrInviteeNotFound
	}
	if invitee.ID == actorUserID {
		return nil, domainerrors.ErrCannotInviteSelf
	}

	existing, err := uc.workspaceRepo.FindMember(ctx, workspaceID, invitee.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domainerrors.ErrAlreadyMember
	}

	now := time.Now().UTC()
	member := &entities.WorkspaceMember{
		ID:          uuid.NewString(),
		WorkspaceID: workspaceID,
		UserID:      invitee.ID,
		Role:        role,
		JoinedAt:    now,
	}
	if err := uc.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("add invited member: %w", err)
	}
	logArchiveAction(
		ctx, uc.auditRepo, actorUserID,
		authdomain.AuditActionArchiveMemberInvited,
		auditResourceWorkspaceMember, invitee.ID, "Archive workspace member invited successfully",
	)
	return &dtos.WorkspaceMemberResponse{
		UserID:      invitee.ID,
		Email:       invitee.Email,
		DisplayName: invitee.DisplayName,
		Role:        role,
		JoinedAt:    now,
	}, nil
}

// UpdateWorkspaceMemberRoleUseCase changes a non-owner member role.
type UpdateWorkspaceMemberRoleUseCase struct {
	workspaceRepo ports.WorkspaceRepository
	auditRepo     authports.AuditRepository
}

// NewUpdateWorkspaceMemberRoleUseCase creates the use case.
func NewUpdateWorkspaceMemberRoleUseCase(
	workspaceRepo ports.WorkspaceRepository,
	auditRepo authports.AuditRepository,
) *UpdateWorkspaceMemberRoleUseCase {
	return &UpdateWorkspaceMemberRoleUseCase{workspaceRepo: workspaceRepo, auditRepo: auditRepo}
}

// Execute updates the role.
func (uc *UpdateWorkspaceMemberRoleUseCase) Execute(
	ctx context.Context,
	actorUserID, workspaceID, targetUserID, role string,
) (*dtos.WorkspaceMemberResponse, error) {
	if err := requireOwner(ctx, uc.workspaceRepo, actorUserID, workspaceID); err != nil {
		return nil, err
	}
	targetUserID = strings.TrimSpace(targetUserID)
	if targetUserID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}
	role = strings.TrimSpace(role)
	if role != entities.WorkspaceRoleMember && role != entities.WorkspaceRoleViewer {
		return nil, domainerrors.ErrInvalidMemberRole
	}

	target, err := uc.workspaceRepo.FindMember(ctx, workspaceID, targetUserID)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, domainerrors.ErrNotWorkspaceMember
	}
	if target.Role == entities.WorkspaceRoleOwner {
		return nil, domainerrors.ErrCannotModifyOwner
	}
	if err := uc.workspaceRepo.UpdateMemberRole(ctx, workspaceID, targetUserID, role); err != nil {
		return nil, err
	}
	logArchiveAction(
		ctx, uc.auditRepo, actorUserID,
		authdomain.AuditActionArchiveMemberRoleUpdated,
		auditResourceWorkspaceMember, targetUserID, "Archive workspace member role updated successfully",
	)
	target.Role = role
	rows, err := uc.workspaceRepo.ListMembers(ctx, workspaceID)
	if err != nil {
		return &dtos.WorkspaceMemberResponse{UserID: targetUserID, Role: role, JoinedAt: target.JoinedAt}, nil
	}
	for i := range rows {
		if rows[i].UserID == targetUserID {
			resp := toMemberResponse(&rows[i])
			return &resp, nil
		}
	}
	return &dtos.WorkspaceMemberResponse{UserID: targetUserID, Role: role, JoinedAt: target.JoinedAt}, nil
}

// RemoveWorkspaceMemberUseCase removes a non-owner member.
type RemoveWorkspaceMemberUseCase struct {
	workspaceRepo ports.WorkspaceRepository
	auditRepo     authports.AuditRepository
}

// NewRemoveWorkspaceMemberUseCase creates the use case.
func NewRemoveWorkspaceMemberUseCase(
	workspaceRepo ports.WorkspaceRepository,
	auditRepo authports.AuditRepository,
) *RemoveWorkspaceMemberUseCase {
	return &RemoveWorkspaceMemberUseCase{workspaceRepo: workspaceRepo, auditRepo: auditRepo}
}

// Execute removes the member.
func (uc *RemoveWorkspaceMemberUseCase) Execute(ctx context.Context, actorUserID, workspaceID, targetUserID string) error {
	if err := requireOwner(ctx, uc.workspaceRepo, actorUserID, workspaceID); err != nil {
		return err
	}
	targetUserID = strings.TrimSpace(targetUserID)
	if targetUserID == "" {
		return domainerrors.ErrUserIDRequired
	}
	target, err := uc.workspaceRepo.FindMember(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return domainerrors.ErrNotWorkspaceMember
	}
	if target.Role == entities.WorkspaceRoleOwner {
		return domainerrors.ErrCannotModifyOwner
	}
	if err := uc.workspaceRepo.RemoveMember(ctx, workspaceID, targetUserID); err != nil {
		return err
	}
	logArchiveAction(
		ctx, uc.auditRepo, actorUserID,
		authdomain.AuditActionArchiveMemberRemoved,
		auditResourceWorkspaceMember, targetUserID, "Archive workspace member removed successfully",
	)
	return nil
}

func requireMembership(
	ctx context.Context,
	repo ports.WorkspaceRepository,
	userID, workspaceID string,
) (*entities.WorkspaceMember, error) {
	userID = strings.TrimSpace(userID)
	workspaceID = strings.TrimSpace(workspaceID)
	if userID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}
	if workspaceID == "" {
		return nil, domainerrors.ErrWorkspaceIDRequired
	}
	ws, err := repo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, domainerrors.ErrWorkspaceNotFound
	}
	member, err := repo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, domainerrors.ErrNotWorkspaceMember
	}
	return member, nil
}

func requireOwner(
	ctx context.Context,
	repo ports.WorkspaceRepository,
	userID, workspaceID string,
) error {
	member, err := requireMembership(ctx, repo, userID, workspaceID)
	if err != nil {
		return err
	}
	if member.Role != entities.WorkspaceRoleOwner {
		return domainerrors.ErrInsufficientWorkspaceRole
	}
	return nil
}

func toMemberResponse(m *entities.WorkspaceMemberDetail) dtos.WorkspaceMemberResponse {
	name := strings.TrimSpace(m.FirstName + " " + m.LastName)
	if name == "" {
		name = m.Email
	}
	return dtos.WorkspaceMemberResponse{
		UserID:      m.UserID,
		Email:       m.Email,
		DisplayName: name,
		Role:        m.Role,
		JoinedAt:    m.JoinedAt,
	}
}
