package dtos

import "time"

// WorkspaceResponse is the API/view DTO for a workspace.
type WorkspaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	OwnerUserID string    `json:"owner_user_id"`
	IsActive    bool      `json:"is_active"`
	MemberRole  string    `json:"member_role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateHouseholdRequest creates a household archive.
type CreateHouseholdRequest struct {
	Name string `json:"name"`
}

// InviteMemberRequest invites an existing user by email.
type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"` // member | viewer
}

// UpdateMemberRoleRequest changes a member role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

// WorkspaceMemberResponse is a member row for API/views.
type WorkspaceMemberResponse struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}
