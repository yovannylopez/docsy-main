package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

// WorkspaceRepositoryAdapter implements WorkspaceRepository with sqlx.
type WorkspaceRepositoryAdapter struct {
	db *sqlx.DB
}

// NewWorkspaceRepositoryAdapter creates the adapter.
func NewWorkspaceRepositoryAdapter(db *sqlx.DB) *WorkspaceRepositoryAdapter {
	return &WorkspaceRepositoryAdapter{db: db}
}

// FindPersonalByOwner returns the personal workspace for the owner, or nil.
func (r *WorkspaceRepositoryAdapter) FindPersonalByOwner(ctx context.Context, ownerUserID string) (*entities.Workspace, error) {
	const q = `
		SELECT id, name, type, owner_user_id, is_active, created_at, updated_at
		FROM archive_workspaces
		WHERE owner_user_id = $1 AND type = $2 AND is_active = true
		LIMIT 1`

	var ws entities.Workspace
	err := r.db.GetContext(ctx, &ws, q, ownerUserID, entities.WorkspaceTypePersonal)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query personal workspace: %w", err)
	}
	return &ws, nil
}

// FindByID returns a workspace by id or nil.
func (r *WorkspaceRepositoryAdapter) FindByID(ctx context.Context, workspaceID string) (*entities.Workspace, error) {
	const q = `
		SELECT id, name, type, owner_user_id, is_active, created_at, updated_at
		FROM archive_workspaces
		WHERE id = $1 AND is_active = true
		LIMIT 1`

	var ws entities.Workspace
	err := r.db.GetContext(ctx, &ws, q, workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query workspace: %w", err)
	}
	return &ws, nil
}

// ListForUser returns workspaces the user belongs to.
func (r *WorkspaceRepositoryAdapter) ListForUser(ctx context.Context, userID string) ([]entities.WorkspaceMembership, error) {
	const q = `
		SELECT w.id, w.name, w.type, w.owner_user_id, w.is_active, w.created_at, w.updated_at, m.role AS member_role
		FROM archive_workspaces w
		INNER JOIN archive_workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id = $1 AND w.is_active = true
		ORDER BY CASE w.type WHEN 'personal' THEN 0 WHEN 'household' THEN 1 ELSE 2 END, w.name ASC`

	var rows []entities.WorkspaceMembership
	if err := r.db.SelectContext(ctx, &rows, q, userID); err != nil {
		return nil, fmt.Errorf("list workspaces for user: %w", err)
	}
	return rows, nil
}

// CreateWorkspace inserts a workspace row.
func (r *WorkspaceRepositoryAdapter) CreateWorkspace(ctx context.Context, workspace *entities.Workspace) error {
	const q = `
		INSERT INTO archive_workspaces (id, name, type, owner_user_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, q,
		workspace.ID,
		workspace.Name,
		workspace.Type,
		workspace.OwnerUserID,
		workspace.IsActive,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert workspace: %w", err)
	}
	return nil
}

// AddMember inserts a workspace membership.
func (r *WorkspaceRepositoryAdapter) AddMember(ctx context.Context, member *entities.WorkspaceMember) error {
	const q = `
		INSERT INTO archive_workspace_members (id, workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, q,
		member.ID,
		member.WorkspaceID,
		member.UserID,
		member.Role,
		member.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("insert workspace member: %w", err)
	}
	return nil
}

// FindMember returns a membership or nil.
func (r *WorkspaceRepositoryAdapter) FindMember(ctx context.Context, workspaceID, userID string) (*entities.WorkspaceMember, error) {
	const q = `
		SELECT id, workspace_id, user_id, role, joined_at
		FROM archive_workspace_members
		WHERE workspace_id = $1 AND user_id = $2
		LIMIT 1`

	var m entities.WorkspaceMember
	err := r.db.GetContext(ctx, &m, q, workspaceID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query workspace member: %w", err)
	}
	return &m, nil
}

// ListMembers returns members with basic user profile fields.
func (r *WorkspaceRepositoryAdapter) ListMembers(ctx context.Context, workspaceID string) ([]entities.WorkspaceMemberDetail, error) {
	const q = `
		SELECT m.id, m.workspace_id, m.user_id, m.role, m.joined_at,
		       u.email, COALESCE(u.first_name, '') AS first_name, COALESCE(u.last_name, '') AS last_name
		FROM archive_workspace_members m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = $1
		ORDER BY CASE m.role WHEN 'owner' THEN 0 WHEN 'member' THEN 1 ELSE 2 END, u.email ASC`

	var rows []entities.WorkspaceMemberDetail
	if err := r.db.SelectContext(ctx, &rows, q, workspaceID); err != nil {
		return nil, fmt.Errorf("list workspace members: %w", err)
	}
	return rows, nil
}

// UpdateMemberRole updates a member's role.
func (r *WorkspaceRepositoryAdapter) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	const q = `UPDATE archive_workspace_members SET role = $3 WHERE workspace_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, q, workspaceID, userID, role)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("update member role: no rows affected")
	}
	return nil
}

// RemoveMember deletes a membership.
func (r *WorkspaceRepositoryAdapter) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	const q = `DELETE FROM archive_workspace_members WHERE workspace_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, q, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("remove workspace member: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("remove workspace member: no rows affected")
	}
	return nil
}
