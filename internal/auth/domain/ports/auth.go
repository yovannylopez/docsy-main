package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// AuthenticationService defines the interface for authentication services
type AuthenticationService interface {
	Authenticate(ctx context.Context, username, password string) (*entities.AuthToken, error)
	ValidateToken(ctx context.Context, token string) (*entities.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entities.AuthToken, error)
	// New methods for the robust model
	Login(
		ctx context.Context,
		email, password, userAgent, ipAddress string,
	) (*entities.AuthToken, *entities.Session, error)
	// Logout revokes the session from the access JWT (claim session_id) or can be extended in the handler with refresh in body.
	Logout(ctx context.Context, accessToken, userAgent, ipAddress string) error
	LogoutAll(ctx context.Context, userID string) error
	ValidateSession(ctx context.Context, sessionID string) (*entities.Session, error)
	RevokeSession(ctx context.Context, sessionID string, reason string) error
}

// AuthorizationService defines the interface for authorization services
type AuthorizationService interface {
	HasPermission(ctx context.Context, user *entities.User, resource, action string) bool
	GetUserPermissions(ctx context.Context, user *entities.User) ([]string, error)
	// New methods for the robust model
	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
	GetUserPermissionsDetailed(ctx context.Context, userID string) ([]entities.Permission, error)
	GetRolePermissions(ctx context.Context, roleID string) ([]entities.Permission, error)
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
}

// SecurityService defines the interface for security services
type SecurityService interface {
	ValidatePasswordPolicy(ctx context.Context, password string) error
	CheckPasswordHistory(ctx context.Context, userID, passwordHash string) error
	LockUserAccount(ctx context.Context, userID string, reason string) error
	UnlockUserAccount(ctx context.Context, userID string) error
	CheckAccountLockout(ctx context.Context, userID string) error
	GenerateMFASecret(ctx context.Context, userID string) (string, error)
	ValidateMFAToken(ctx context.Context, userID, token string) (bool, error)
	GenerateVerificationToken(ctx context.Context, userID, tokenType string) (string, error)
	ValidateVerificationToken(ctx context.Context, tokenHash, tokenType string) (*entities.VerificationToken, error)
}

// SessionService defines the interface for session services
type SessionService interface {
	CreateSession(ctx context.Context, userID, userAgent, ipAddress string) (*entities.Session, error)
	ValidateSession(ctx context.Context, sessionID string) (*entities.Session, error)
	RevokeSession(ctx context.Context, sessionID string, reason string) error
	GetUserSessions(ctx context.Context, userID string) ([]entities.Session, error)
	CleanExpiredSessions(ctx context.Context) error
	UpdateSessionActivity(ctx context.Context, sessionID string) error
}

// AuditService defines the interface for audit services
type AuditService interface {
	LogUserAction(ctx context.Context, userID, action, resource, resourceID string) error
	LogSystemAction(ctx context.Context, action, resource, resourceID string) error
	LogLoginAttempt(ctx context.Context, userID, email, ipAddress, userAgent string, success bool) error
	LogPasswordChange(ctx context.Context, userID, changedBy string) error
	LogRoleAssignment(ctx context.Context, userID, roleID, assignedBy string) error
	LogPermissionGrant(ctx context.Context, roleID, permissionID, grantedBy string) error
	GetUserAuditLogs(ctx context.Context, userID string, limit, offset int) ([]entities.AuditLog, error)
	GetSystemAuditLogs(ctx context.Context, limit, offset int) ([]entities.AuditLog, error)
}

// SystemConfigService defines the interface for system configuration services
type SystemConfigService interface {
	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key, value, description string, updatedBy string) error
	GetAllConfig(ctx context.Context) (map[string]string, error)
	ValidatePasswordPolicy(ctx context.Context, password string) error
	GetPasswordPolicy(ctx context.Context) (map[string]any, error)
	GetSessionPolicy(ctx context.Context) (map[string]any, error)
	GetLockoutPolicy(ctx context.Context) (map[string]any, error)
}
