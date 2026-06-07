package entities

import (
	"time"
)

// User represents a user in the application (simplified version for the auth module)
type User struct {
	ID                   string     `json:"id"`
	Email                string     `json:"email"`
	Username             *string    `json:"username,omitempty" db:"username"`
	PasswordHash         string     `json:"-" db:"password_hash"`
	FirstName            string     `json:"first_name" db:"first_name"`
	LastName             string     `json:"last_name" db:"last_name"`
	IdentificationNumber *string    `json:"identification_number,omitempty" db:"identification_number"`
	IdentificationType   *string    `json:"identification_type,omitempty" db:"identification_type"`
	Phone                *string    `json:"phone,omitempty"`
	IsActive             bool       `json:"is_active" db:"is_active"`
	IsVerified           bool       `json:"is_verified" db:"is_verified"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	FailedLoginAttempts  int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LastFailedLoginAt    *time.Time `json:"last_failed_login_at,omitempty" db:"last_failed_login_at"`
	LockedUntil          *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	MFAEnabled           bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret            *string    `json:"mfa_secret,omitempty" db:"mfa_secret"`
	PasswordChangedAt    time.Time  `json:"password_changed_at" db:"password_changed_at"`
	MustChangePassword   bool       `json:"must_change_password" db:"must_change_password"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy            *string    `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy            *string    `json:"updated_by,omitempty" db:"updated_by"`
	// Roles will be managed separately
	Roles []Role `json:"roles,omitempty"`
	// PermissionNames added at runtime (RBAC); not persisted in users.
	PermissionNames []string `json:"-" db:"-"`
}

// Role represents a role in the application
type Role struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	IsSystemRole bool      `json:"is_system_role" db:"is_system_role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Permission represents a permission in the application
type Permission struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Resource           string    `json:"resource"`
	Action             string    `json:"action"`
	Description        *string   `json:"description,omitempty"`
	IsSystemPermission bool      `json:"is_system_permission" db:"is_system_permission"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// Session represents a user session
type Session struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id" db:"user_id"`
	RefreshTokenHash  string     `json:"refresh_token_hash" db:"refresh_token_hash"`
	AccessTokenJTI    *string    `json:"access_token_jti,omitempty" db:"access_token_jti"`
	UserAgent         *string    `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress         *string    `json:"ip_address,omitempty" db:"ip_address"`
	Location          *string    `json:"location,omitempty" db:"location"`
	DeviceFingerprint *string    `json:"device_fingerprint,omitempty" db:"device_fingerprint"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt        time.Time  `json:"last_used_at" db:"last_used_at"`
	ExpiresAt         time.Time  `json:"expires_at" db:"expires_at"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	RevokedReason     *string    `json:"revoked_reason,omitempty" db:"revoked_reason"`
}

// PasswordHistory represents the password history of a user
type PasswordHistory struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	ChangedAt    time.Time `json:"changed_at" db:"changed_at"`
	ChangedBy    *string   `json:"changed_by,omitempty" db:"changed_by"`
}

// VerificationToken represents a verification token
type VerificationToken struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id" db:"user_id"`
	TokenHash string `json:"token_hash" db:"token_hash"`
	// TokenType: MFA values are domain.VerificationTokenTypeMFASetup and domain.VerificationTokenTypeMFAChallenge
	// (domain/verification_token.go). Other persisted values include email_verification and password_reset.
	TokenType string     `json:"token_type" db:"token_type"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog represents an audit log
type AuditLog struct {
	ID         string  `json:"id"`
	UserID     *string `json:"user_id,omitempty" db:"user_id"`
	SessionID  *string `json:"session_id,omitempty" db:"session_id"`
	Action     string  `json:"action"`
	Resource   *string `json:"resource,omitempty"`
	ResourceID *string `json:"resource_id,omitempty" db:"resource_id"`
	// Result matches domain.AuditResultSuccess, domain.AuditResultFailure, domain.AuditResultError (domain/audit.go).
	Result        string          `json:"result"`
	Message       *string         `json:"message,omitempty"`
	IPAddress     *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string         `json:"user_agent,omitempty" db:"user_agent"`
	RequestID     *string         `json:"request_id,omitempty" db:"request_id"`
	PreviousData  *map[string]any `json:"previous_data,omitempty" db:"previous_data"`
	NewData       *map[string]any `json:"new_data,omitempty" db:"new_data"`
	ChangedFields []string        `json:"changed_fields,omitempty" db:"changed_fields"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// SystemConfig represents a system configuration
type SystemConfig struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description *string   `json:"description,omitempty"`
	IsSensitive bool      `json:"is_sensitive" db:"is_sensitive"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy   *string   `json:"updated_by,omitempty" db:"updated_by"`
}

// AuthToken represents an authentication token
type AuthToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}
