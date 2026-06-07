package ports

import (
	"context"
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// UserRepository defines the user repository operations for authentication
type UserRepository interface {
	// Basic authentication operations
	Create(ctx context.Context, user *entities.User) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, userID string) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	// List and search operations (consumed via port by other contexts, e.g. users)
	GetAllUsers(ctx context.Context, limit, offset int) ([]entities.User, error)
	GetTotalUsersCount(ctx context.Context) (int, error)
	SearchUsers(ctx context.Context, query string, activo *bool, limit, offset int) ([]entities.User, error)
	CountSearchUsers(ctx context.Context, query string, activo *bool) (int, error)
	// Role operations (required for authentication)
	GetRoleByName(ctx context.Context, roleName string) (*entities.Role, error)
	// Security operations
	UpdateLastLogin(ctx context.Context, userID string) error
	IncrementFailedLoginAttempts(ctx context.Context, userID string) error
	// RecordFailedPasswordAttempt increments failed_login_attempts and, when maxAttempts > 0
	// and the new count reaches the threshold, sets locked_until to now + lockDuration (atomic).
	// Pass maxAttempts 0 to only increment (same effect as IncrementFailedLoginAttempts).
	RecordFailedPasswordAttempt(
		ctx context.Context, userID string, maxAttempts int, lockDuration time.Duration,
	) (FailedPasswordAttemptResult, error)
	ResetFailedLoginAttempts(ctx context.Context, userID string) error
	LockUserAccount(ctx context.Context, userID string, until *time.Time) error
	UnlockUserAccount(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	SetMustChangePassword(ctx context.Context, userID string, mustChange bool) error
	// UpdateMFASecret stores the encrypted secret without flipping mfa_enabled.
	// Used during the MFA setup flow to persist the pending secret before confirmation.
	UpdateMFASecret(ctx context.Context, userID, encryptedSecret string) error
	EnableMFA(ctx context.Context, userID, secret string) error
	DisableMFA(ctx context.Context, userID string) error
	VerifyUser(ctx context.Context, userID string) error
}

// FailedPasswordAttemptResult is the DB state after recording a failed local password login.
type FailedPasswordAttemptResult struct {
	FailedAttempts int
	LockedUntil    *time.Time
}

// PasswordHistoryRepository defines the interface for the password history repository
type PasswordHistoryRepository interface {
	Create(ctx context.Context, passwordHistory *entities.PasswordHistory) error
	GetUserPasswordHistory(ctx context.Context, userID string, limit int) ([]entities.PasswordHistory, error)
	CheckPasswordInHistory(ctx context.Context, userID, passwordHash string) (bool, error)
	CleanOldPasswordHistory(ctx context.Context, userID string, keepCount int) error
}

// VerificationTokenRepository defines the interface for the verification token repository
type VerificationTokenRepository interface {
	CreateToken(ctx context.Context, token *entities.VerificationToken) error
	FindTokenByHash(ctx context.Context, tokenHash string) (*entities.VerificationToken, error)
	MarkTokenAsUsed(ctx context.Context, tokenID string) error
	CleanExpiredTokens(ctx context.Context) error
	GetUserTokens(ctx context.Context, userID string, tokenType string) ([]entities.VerificationToken, error)
}

// PermissionRepository defines the interface for the permissions repository
type PermissionRepository interface {
	GetPermissionByName(ctx context.Context, name string) (*entities.Permission, error)
	GetPermissionByResourceAction(ctx context.Context, resource, action string) (*entities.Permission, error)
	GetAllPermissions(ctx context.Context) ([]entities.Permission, error)
	GetUserPermissions(ctx context.Context, userID string) ([]entities.Permission, error)
	GetRolePermissions(ctx context.Context, roleID string) ([]entities.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID, grantedBy string) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
}
