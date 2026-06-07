package dtos

// UserResponse represents the user response without sensitive information
type UserResponse struct {
	ID                  string         `json:"id"`
	Email               string         `json:"email"`
	Username            *string        `json:"username,omitempty"`
	FirstName           string         `json:"first_name"`
	LastName            string         `json:"last_name"`
	Phone               *string        `json:"phone,omitempty"`
	IsActive            bool           `json:"is_active"`
	IsVerified          bool           `json:"is_verified"`
	LastLoginAt         *string        `json:"last_login_at,omitempty"`
	FailedLoginAttempts int            `json:"failed_login_attempts"`
	LastFailedLoginAt   *string        `json:"last_failed_login_at,omitempty"`
	LockedUntil         *string        `json:"locked_until,omitempty"`
	MFAEnabled          bool           `json:"mfa_enabled"`
	PasswordChangedAt   string         `json:"password_changed_at"`
	MustChangePassword  bool           `json:"must_change_password"`
	CreatedAt           string         `json:"created_at"`
	UpdatedAt           string         `json:"updated_at"`
	Roles               []RoleResponse `json:"roles"`
}

// RoleResponse represents the role response
type RoleResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	IsSystemRole bool    `json:"is_system_role"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// PermissionResponse represents the permission response
type PermissionResponse struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Resource           string  `json:"resource"`
	Action             string  `json:"action"`
	Description        *string `json:"description,omitempty"`
	IsSystemPermission bool    `json:"is_system_permission"`
	CreatedAt          string  `json:"created_at"`
}

// SessionResponse represents the session response
type SessionResponse struct {
	ID                string  `json:"id"`
	UserID            string  `json:"user_id"`
	AccessTokenJTI    *string `json:"access_token_jti,omitempty"`
	UserAgent         *string `json:"user_agent,omitempty"`
	IPAddress         *string `json:"ip_address,omitempty"`
	Location          *string `json:"location,omitempty"`
	DeviceFingerprint *string `json:"device_fingerprint,omitempty"`
	CreatedAt         string  `json:"created_at"`
	LastUsedAt        string  `json:"last_used_at"`
	ExpiresAt         string  `json:"expires_at"`
	IsActive          bool    `json:"is_active"`
	RevokedAt         *string `json:"revoked_at,omitempty"`
	RevokedReason     *string `json:"revoked_reason,omitempty"`
}

// TokenResponse represents the authentication token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    string `json:"expires_at"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// AuditLogResponse represents the audit log response
type AuditLogResponse struct {
	ID         string  `json:"id"`
	UserID     *string `json:"user_id,omitempty"`
	SessionID  *string `json:"session_id,omitempty"`
	Action     string  `json:"action"`
	Resource   *string `json:"resource,omitempty"`
	ResourceID *string `json:"resource_id,omitempty"`
	Result     string  `json:"result"`
	Message    *string `json:"message,omitempty"`
	IPAddress  *string `json:"ip_address,omitempty"`
	UserAgent  *string `json:"user_agent,omitempty"`
	RequestID  *string `json:"request_id,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

// SystemConfigResponse represents the system configuration response
type SystemConfigResponse struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
	IsSensitive bool    `json:"is_sensitive"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	UpdatedBy   *string `json:"updated_by,omitempty"`
}
