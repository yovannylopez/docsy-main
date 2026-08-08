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
)

const defaultPersonalWorkspaceName = "Mi archivo"

// EnsurePersonalWorkspaceUseCase creates a personal workspace on first access.
type EnsurePersonalWorkspaceUseCase struct {
	workspaceRepo ports.WorkspaceRepository
}

// NewEnsurePersonalWorkspaceUseCase creates the use case.
func NewEnsurePersonalWorkspaceUseCase(workspaceRepo ports.WorkspaceRepository) *EnsurePersonalWorkspaceUseCase {
	return &EnsurePersonalWorkspaceUseCase{workspaceRepo: workspaceRepo}
}

// Execute returns the user's personal workspace, creating it if missing.
func (uc *EnsurePersonalWorkspaceUseCase) Execute(ctx context.Context, userID string) (*dtos.WorkspaceResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domainerrors.ErrUserIDRequired
	}

	existing, err := uc.workspaceRepo.FindPersonalByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find personal workspace: %w", err)
	}
	if existing != nil {
		member, mErr := uc.workspaceRepo.FindMember(ctx, existing.ID, userID)
		if mErr != nil {
			return nil, fmt.Errorf("find workspace member: %w", mErr)
		}
		role := entities.WorkspaceRoleOwner
		if member != nil {
			role = member.Role
		}
		return toWorkspaceResponse(existing, role), nil
	}

	now := time.Now().UTC()
	workspace := &entities.Workspace{
		ID:          uuid.NewString(),
		Name:        defaultPersonalWorkspaceName,
		Type:        entities.WorkspaceTypePersonal,
		OwnerUserID: userID,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := uc.workspaceRepo.CreateWorkspace(ctx, workspace); err != nil {
		return nil, fmt.Errorf("create personal workspace: %w", err)
	}

	member := &entities.WorkspaceMember{
		ID:          uuid.NewString(),
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        entities.WorkspaceRoleOwner,
		JoinedAt:    now,
	}
	if err := uc.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("add workspace owner: %w", err)
	}

	return toWorkspaceResponse(workspace, entities.WorkspaceRoleOwner), nil
}

func toWorkspaceResponse(w *entities.Workspace, role string) *dtos.WorkspaceResponse {
	return &dtos.WorkspaceResponse{
		ID:          w.ID,
		Name:        w.Name,
		Type:        w.Type,
		OwnerUserID: w.OwnerUserID,
		IsActive:    w.IsActive,
		MemberRole:  role,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}
