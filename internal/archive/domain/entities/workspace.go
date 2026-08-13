package entities

import "time"

// WorkspaceType identifies the kind of archive container.
const (
	WorkspaceTypePersonal     = "personal"
	WorkspaceTypeHousehold    = "household"
	WorkspaceTypeOrganization = "organization"
)

// WorkspaceMemberRole identifies membership within a workspace.
const (
	WorkspaceRoleOwner  = "owner"
	WorkspaceRoleMember = "member"
	WorkspaceRoleViewer = "viewer"
)

// Workspace is the multi-tenant container for personal/family/organization archives.
type Workspace struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"`
	OwnerUserID string    `json:"owner_user_id" db:"owner_user_id"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// WorkspaceMember links a user to a workspace with a role.
type WorkspaceMember struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Role        string    `json:"role" db:"role"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

// DocumentCategory is a flat classification for archive documents (system or workspace-custom).
type DocumentCategory struct {
	Code        string    `json:"code" db:"code"`
	WorkspaceID *string   `json:"workspace_id,omitempty" db:"workspace_id"`
	LabelES     string    `json:"label_es" db:"label_es"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// WorkspaceMembership is a workspace plus the caller's role (list query).
type WorkspaceMembership struct {
	Workspace
	MemberRole string `db:"member_role"`
}

// WorkspaceMemberDetail is a membership row enriched for listing.
type WorkspaceMemberDetail struct {
	WorkspaceMember
	Email     string `db:"email"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
}
