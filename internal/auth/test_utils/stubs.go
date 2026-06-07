package test_utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// Constants for test values
const (
	// Pagination limits
	DefaultLimit  = 10
	DefaultOffset = 0

	// User and role limits for testing
	MinUserCount = 1
	MaxUserCount = 5
	MinRoleCount = 1
	MaxRoleCount = 4

	// Intermediate limits for testing
	MidUserCount        = 2
	MidRoleCount        = 2
	HighUserCount       = 3
	HighRoleCount       = 3
	MaxUserCountWithMFA = 4

	// Constants for time
	DefaultTokenExpirationHours = 24
	ShortTokenExpirationHours   = 12

	// Constants for login attempts
	MaxFailedLoginAttempts = 5
	MinFailedLoginAttempts = 2

	// Constants for duplicate strings
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	DefaultIPAddress = "192.168.1.100"

	// Constants for user totals
	DefaultUserTotal = 2

	// Constants for test IDs
	NewUserID           = "new-user-123"
	NewUserEmail        = "newuser@example.com"
	NewUserFirstName    = "New"
	NewUserLastName     = "User"
	RoleUserID          = "role-user"
	RoleUserName        = "user"
	RoleUserDescription = "Regular user role"
	NewAccessToken      = "new_access_token_123"
	NewRefreshToken     = "new_refresh_token_123"
)

// AuthTestEntities contains test entities for the auth module
type AuthTestEntities struct {
	ValidUser      *entities.User
	AdminUser      *entities.User
	InactiveUser   *entities.User
	LockedUser     *entities.User
	UnverifiedUser *entities.User
	UserWithMFA    *entities.User
	UserWithPhone  *entities.User
	EmptyUser      *entities.User
	InvalidUser    *entities.User

	ValidRole   entities.Role
	AdminRole   entities.Role
	UserRole    entities.Role
	GuestRole   entities.Role
	EmptyRole   entities.Role
	InvalidRole entities.Role

	ValidSession   *entities.Session
	ExpiredSession *entities.Session
	RevokedSession *entities.Session
	EmptySession   *entities.Session
	InvalidSession *entities.Session

	ValidAuditLog   *entities.AuditLog
	FailedAuditLog  *entities.AuditLog
	EmptyAuditLog   *entities.AuditLog
	InvalidAuditLog *entities.AuditLog

	ValidAuthToken   *entities.AuthToken
	ExpiredAuthToken *entities.AuthToken
	EmptyAuthToken   *entities.AuthToken
	InvalidAuthToken *entities.AuthToken
}

// AuthTestDTOs contains test DTOs for the auth module
type AuthTestDTOs struct {
	ValidLoginRequest   *dtos.LoginRequest
	InvalidLoginRequest *dtos.LoginRequest
	EmptyLoginRequest   *dtos.LoginRequest

	ValidLoginResponse *dtos.LoginResponse
	EmptyLoginResponse *dtos.LoginResponse

	ValidUserResponse   *dtos.UserResponse
	AdminUserResponse   *dtos.UserResponse
	EmptyUserResponse   *dtos.UserResponse
	InvalidUserResponse *dtos.UserResponse

	ValidRoleResponse *dtos.RoleResponse
	AdminRoleResponse *dtos.RoleResponse
	EmptyRoleResponse *dtos.RoleResponse

	ValidSessionResponse   *dtos.SessionResponse
	ExpiredSessionResponse *dtos.SessionResponse
	EmptySessionResponse   *dtos.SessionResponse

	ValidTokenResponse   *dtos.TokenResponse
	ExpiredTokenResponse *dtos.TokenResponse
	EmptyTokenResponse   *dtos.TokenResponse

	ValidAuditLogResponse  *dtos.AuditLogResponse
	FailedAuditLogResponse *dtos.AuditLogResponse
	EmptyAuditLogResponse  *dtos.AuditLogResponse
}

// AuthTestUseCases contains test use cases for the auth module
type AuthTestUseCases struct {
	// GetUsersUseCase was moved to the users module
	// ValidGetUsersRequest   *usecases.GetUsersRequest
	// InvalidGetUsersRequest *usecases.GetUsersRequest
	// EmptyGetUsersRequest   *usecases.GetUsersRequest

	// UsersListResponse was moved to the users module
	// ValidUsersListResponse   *dtos.UsersListResponse
	// EmptyUsersListResponse   *dtos.UsersListResponse
	// InvalidUsersListResponse *dtos.UsersListResponse
}

// AuthTestScenarios contains test scenarios for the auth module
type AuthTestScenarios struct {
	SuccessfulLogin    AuthLoginScenario
	FailedLogin        AuthLoginScenario
	AccountLocked      AuthLoginScenario
	AccountInactive    AuthLoginScenario
	InvalidCredentials AuthLoginScenario

	SuccessfulGetUsers AuthGetUsersScenario
	FailedGetUsers     AuthGetUsersScenario
	EmptyUsers         AuthGetUsersScenario
	PaginationError    AuthGetUsersScenario
}

// AuthLoginScenario represents a login scenario
type AuthLoginScenario struct {
	Request          *dtos.LoginRequest
	User             *entities.User
	Token            *entities.AuthToken
	Session          *entities.Session
	ExpectedResponse *dtos.LoginResponse
	ExpectedError    string
	UserAgent        string
	IPAddress        string
}

// AuthGetUsersScenario represents a get users scenario
// GetUsersUseCase was moved to the users module
type AuthGetUsersScenario struct{}

// AuthStubs contains all test stubs for the auth module
type AuthStubs struct {
	Entities  AuthTestEntities
	DTOs      AuthTestDTOs
	UseCases  AuthTestUseCases
	Scenarios AuthTestScenarios
}

// NewAuthStubs creates a new stubs instance for the auth module
func NewAuthStubs() *AuthStubs {
	return &AuthStubs{
		Entities:  newAuthTestEntities(),
		DTOs:      newAuthTestDTOs(),
		UseCases:  newAuthTestUseCases(),
		Scenarios: newAuthTestScenarios(),
	}
}

// newAuthTestEntities creates test entities for the auth module
func newAuthTestEntities() AuthTestEntities {
	now := time.Now()
	userID := uuid.New().String()
	phone := "+1234567890"
	userAgent := DefaultUserAgent
	ipAddress := DefaultIPAddress
	location := "New York, NY"
	deviceFingerprint := "device-fingerprint-123"

	return AuthTestEntities{
		ValidUser: &entities.User{
			ID:                  userID,
			Email:               "test@example.com",
			PasswordHash:        "hashed_password_123",
			FirstName:           "John",
			LastName:            "Doe",
			Phone:               &phone,
			IsActive:            true,
			IsVerified:          true,
			LastLoginAt:         &now,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			MFASecret:           nil,
			PasswordChangedAt:   now,
			MustChangePassword:  false,
			CreatedAt:           now,
			UpdatedAt:           now,
			Roles: []entities.Role{
				{
					ID:           "role-user",
					Name:         "user",
					Description:  stringPtr("Regular user role"),
					IsSystemRole: true,
					IsActive:     true,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
			},
		},
		AdminUser: &entities.User{
			ID:                  uuid.New().String(),
			Email:               "admin@example.com",
			PasswordHash:        "hashed_admin_password",
			FirstName:           "Admin",
			LastName:            "User",
			Phone:               nil,
			IsActive:            true,
			IsVerified:          true,
			LastLoginAt:         &now,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          true,
			MFASecret:           stringPtr("mfa_secret_123"),
			PasswordChangedAt:   now,
			MustChangePassword:  false,
			CreatedAt:           now,
			UpdatedAt:           now,
			Roles: []entities.Role{
				{
					ID:           "role-admin",
					Name:         "admin",
					Description:  stringPtr("Administrator role"),
					IsSystemRole: true,
					IsActive:     true,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
			},
		},
		InactiveUser: &entities.User{
			ID:                  uuid.New().String(),
			Email:               "inactive@example.com",
			PasswordHash:        "hashed_inactive_password",
			FirstName:           "Inactive",
			LastName:            "User",
			Phone:               nil,
			IsActive:            false,
			IsVerified:          true,
			LastLoginAt:         nil,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			MFASecret:           nil,
			PasswordChangedAt:   now,
			MustChangePassword:  false,
			CreatedAt:           now,
			UpdatedAt:           now,
			Roles:               []entities.Role{},
		},
		LockedUser: &entities.User{
			ID:                  uuid.New().String(),
			Email:               "locked@example.com",
			PasswordHash:        "hashed_locked_password",
			FirstName:           "Locked",
			LastName:            "User",
			Phone:               nil,
			IsActive:            true,
			IsVerified:          true,
			LastLoginAt:         nil,
			FailedLoginAttempts: MaxFailedLoginAttempts,
			LastFailedLoginAt:   &now,
			LockedUntil:         timePtr(now.Add(time.Hour)),
			MFAEnabled:          false,
			MFASecret:           nil,
			PasswordChangedAt:   now,
			MustChangePassword:  false,
			CreatedAt:           now,
			UpdatedAt:           now,
			Roles:               []entities.Role{},
		},
		UnverifiedUser: &entities.User{
			ID:                  uuid.New().String(),
			Email:               "unverified@example.com",
			PasswordHash:        "hashed_unverified_password",
			FirstName:           "Unverified",
			LastName:            "User",
			Phone:               nil,
			IsActive:            true,
			IsVerified:          false,
			LastLoginAt:         nil,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			MFASecret:           nil,
			PasswordChangedAt:   now,
			MustChangePassword:  false,
			CreatedAt:           now,
			UpdatedAt:           now,
			Roles:               []entities.Role{},
		},
		UserWithMFA: &entities.User{
			ID:                  uuid.New().String(),
			Email:               "mfa@example.com",
			PasswordHash:        "hashed_mfa_password",
			FirstName:           "MFA",
			LastName:            "User",
			Phone:               nil,
			IsActive:            true,
			IsVerified:          true,
			LastLoginAt:         &now,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          true,
			MFASecret:           stringPtr("mfa_secret_456"),
			PasswordChangedAt:   now,
			MustChangePassword:  false,
			CreatedAt:           now,
			UpdatedAt:           now,
			Roles: []entities.Role{
				{
					ID:           "role-user",
					Name:         "user",
					Description:  stringPtr("Regular user role"),
					IsSystemRole: true,
					IsActive:     true,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
			},
		},
		UserWithPhone: &entities.User{
			ID:                  uuid.New().String(),
			Email:               "phone@example.com",
			PasswordHash:        "hashed_phone_password",
			FirstName:           "Phone",
			LastName:            "User",
			Phone:               &phone,
			IsActive:            true,
			IsVerified:          true,
			LastLoginAt:         &now,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			MFASecret:           nil,
			PasswordChangedAt:   now,
			MustChangePassword:  false,
			CreatedAt:           now,
			UpdatedAt:           now,
			Roles: []entities.Role{
				{
					ID:           "role-user",
					Name:         "user",
					Description:  stringPtr("Regular user role"),
					IsSystemRole: true,
					IsActive:     true,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
			},
		},
		EmptyUser: &entities.User{
			ID:                  "",
			Email:               "",
			PasswordHash:        "",
			FirstName:           "",
			LastName:            "",
			Phone:               nil,
			IsActive:            false,
			IsVerified:          false,
			LastLoginAt:         nil,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			MFASecret:           nil,
			PasswordChangedAt:   time.Time{},
			MustChangePassword:  false,
			CreatedAt:           time.Time{},
			UpdatedAt:           time.Time{},
			Roles:               nil,
		},
		InvalidUser: &entities.User{
			ID:                  "invalid-id",
			Email:               "invalid-email",
			PasswordHash:        "",
			FirstName:           "",
			LastName:            "",
			Phone:               nil,
			IsActive:            false,
			IsVerified:          false,
			LastLoginAt:         nil,
			FailedLoginAttempts: -1,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			MFASecret:           nil,
			PasswordChangedAt:   time.Time{},
			MustChangePassword:  false,
			CreatedAt:           time.Time{},
			UpdatedAt:           time.Time{},
			Roles:               nil,
		},
		ValidRole: entities.Role{
			ID:           "role-user",
			Name:         "user",
			Description:  stringPtr("Regular user role"),
			IsSystemRole: true,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		AdminRole: entities.Role{
			ID:           "role-admin",
			Name:         "admin",
			Description:  stringPtr("Administrator role"),
			IsSystemRole: true,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		UserRole: entities.Role{
			ID:           "role-user",
			Name:         "user",
			Description:  stringPtr("Regular user role"),
			IsSystemRole: true,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		GuestRole: entities.Role{
			ID:           "role-guest",
			Name:         "guest",
			Description:  stringPtr("Guest role"),
			IsSystemRole: true,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		EmptyRole: entities.Role{
			ID:           "",
			Name:         "",
			Description:  nil,
			IsSystemRole: false,
			IsActive:     false,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
		},
		InvalidRole: entities.Role{
			ID:           "invalid-role",
			Name:         "",
			Description:  nil,
			IsSystemRole: false,
			IsActive:     false,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
		},
		ValidSession: &entities.Session{
			ID:                uuid.New().String(),
			UserID:            userID,
			RefreshTokenHash:  "hashed_refresh_token_123",
			AccessTokenJTI:    stringPtr("jti_123"),
			UserAgent:         &userAgent,
			IPAddress:         &ipAddress,
			Location:          &location,
			DeviceFingerprint: &deviceFingerprint,
			CreatedAt:         now,
			LastUsedAt:        now,
			ExpiresAt:         now.Add(DefaultTokenExpirationHours * time.Hour),
			IsActive:          true,
			RevokedAt:         nil,
			RevokedReason:     nil,
		},
		ExpiredSession: &entities.Session{
			ID:                uuid.New().String(),
			UserID:            userID,
			RefreshTokenHash:  "hashed_expired_refresh_token",
			AccessTokenJTI:    stringPtr("jti_expired"),
			UserAgent:         &userAgent,
			IPAddress:         &ipAddress,
			Location:          &location,
			DeviceFingerprint: &deviceFingerprint,
			CreatedAt:         now.Add(-24 * time.Hour),
			LastUsedAt:        now.Add(-1 * time.Hour),
			ExpiresAt:         now.Add(-1 * time.Hour),
			IsActive:          false,
			RevokedAt:         nil,
			RevokedReason:     nil,
		},
		RevokedSession: &entities.Session{
			ID:                uuid.New().String(),
			UserID:            userID,
			RefreshTokenHash:  "hashed_revoked_refresh_token",
			AccessTokenJTI:    stringPtr("jti_revoked"),
			UserAgent:         &userAgent,
			IPAddress:         &ipAddress,
			Location:          &location,
			DeviceFingerprint: &deviceFingerprint,
			CreatedAt:         now.Add(-ShortTokenExpirationHours * time.Hour),
			LastUsedAt:        now.Add(-1 * time.Hour),
			ExpiresAt:         now.Add(ShortTokenExpirationHours * time.Hour),
			IsActive:          false,
			RevokedAt:         timePtr(now.Add(-1 * time.Hour)),
			RevokedReason:     stringPtr("User logout"),
		},
		EmptySession: &entities.Session{
			ID:                "",
			UserID:            "",
			RefreshTokenHash:  "",
			AccessTokenJTI:    nil,
			UserAgent:         nil,
			IPAddress:         nil,
			Location:          nil,
			DeviceFingerprint: nil,
			CreatedAt:         time.Time{},
			LastUsedAt:        time.Time{},
			ExpiresAt:         time.Time{},
			IsActive:          false,
			RevokedAt:         nil,
			RevokedReason:     nil,
		},
		InvalidSession: &entities.Session{
			ID:                "invalid-session-id",
			UserID:            "invalid-user-id",
			RefreshTokenHash:  "",
			AccessTokenJTI:    nil,
			UserAgent:         nil,
			IPAddress:         nil,
			Location:          nil,
			DeviceFingerprint: nil,
			CreatedAt:         time.Time{},
			LastUsedAt:        time.Time{},
			ExpiresAt:         time.Time{},
			IsActive:          false,
			RevokedAt:         nil,
			RevokedReason:     nil,
		},
		ValidAuditLog: &entities.AuditLog{
			ID:         uuid.New().String(),
			UserID:     &userID,
			SessionID:  stringPtr(uuid.New().String()),
			Action:     "user.login",
			Resource:   stringPtr("auth"),
			ResourceID: stringPtr(userID),
			Result:     domain.AuditResultSuccess,
			Message:    stringPtr("User logged in successfully"),
			IPAddress:  &ipAddress,
			UserAgent:  &userAgent,
			RequestID:  stringPtr(uuid.New().String()),
			CreatedAt:  now,
		},
		FailedAuditLog: &entities.AuditLog{
			ID:         uuid.New().String(),
			UserID:     nil,
			SessionID:  nil,
			Action:     "user.login",
			Resource:   stringPtr("auth"),
			ResourceID: nil,
			Result:     domain.AuditResultFailure,
			Message:    stringPtr("Invalid credentials"),
			IPAddress:  &ipAddress,
			UserAgent:  &userAgent,
			RequestID:  stringPtr(uuid.New().String()),
			CreatedAt:  now,
		},
		EmptyAuditLog: &entities.AuditLog{
			ID:         "",
			UserID:     nil,
			SessionID:  nil,
			Action:     "",
			Resource:   nil,
			ResourceID: nil,
			Result:     "",
			Message:    nil,
			IPAddress:  nil,
			UserAgent:  nil,
			RequestID:  nil,
			CreatedAt:  time.Time{},
		},
		InvalidAuditLog: &entities.AuditLog{
			ID:         "invalid-audit-id",
			UserID:     stringPtr("invalid-user-id"),
			SessionID:  stringPtr("invalid-session-id"),
			Action:     "",
			Resource:   nil,
			ResourceID: nil,
			Result:     "",
			Message:    nil,
			IPAddress:  nil,
			UserAgent:  nil,
			RequestID:  nil,
			CreatedAt:  time.Time{},
		},
		ValidAuthToken: &entities.AuthToken{
			AccessToken:  "valid_access_token_123",
			TokenType:    "Bearer",
			ExpiresAt:    now.Add(time.Hour),
			RefreshToken: "valid_refresh_token_123",
		},
		ExpiredAuthToken: &entities.AuthToken{
			AccessToken:  "expired_access_token_123",
			TokenType:    "Bearer",
			ExpiresAt:    now.Add(-time.Hour),
			RefreshToken: "expired_refresh_token_123",
		},
		EmptyAuthToken: &entities.AuthToken{
			AccessToken:  "",
			TokenType:    "",
			ExpiresAt:    time.Time{},
			RefreshToken: "",
		},
		InvalidAuthToken: &entities.AuthToken{
			AccessToken:  "invalid_token",
			TokenType:    "",
			ExpiresAt:    time.Time{},
			RefreshToken: "",
		},
	}
}

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// timePtr creates a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}

// newAuthTestDTOs creates test DTOs for the auth module
func newAuthTestDTOs() AuthTestDTOs {
	now := time.Now()
	phone := "+1234567890"
	userAgent := DefaultUserAgent
	ipAddress := DefaultIPAddress
	location := "New York, NY"
	deviceFingerprint := "device-fingerprint-123"

	return AuthTestDTOs{
		ValidLoginRequest: &dtos.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		},
		InvalidLoginRequest: &dtos.LoginRequest{
			Email:    "invalid-email",
			Password: "",
		},
		EmptyLoginRequest: &dtos.LoginRequest{
			Email:    "",
			Password: "",
		},
		ValidLoginResponse: &dtos.LoginResponse{
			User: &dtos.UserResponse{
				ID:                  "user-123",
				Email:               "test@example.com",
				FirstName:           "John",
				LastName:            "Doe",
				Phone:               &phone,
				IsActive:            true,
				IsVerified:          true,
				LastLoginAt:         stringPtr(now.Format(time.RFC3339)),
				FailedLoginAttempts: 0,
				LastFailedLoginAt:   nil,
				LockedUntil:         nil,
				MFAEnabled:          false,
				PasswordChangedAt:   now.Format(time.RFC3339),
				MustChangePassword:  false,
				CreatedAt:           now.Format(time.RFC3339),
				UpdatedAt:           now.Format(time.RFC3339),
				Roles: []dtos.RoleResponse{
					{
						ID:           "role-user",
						Name:         "user",
						Description:  stringPtr("Regular user role"),
						IsSystemRole: true,
						IsActive:     true,
						CreatedAt:    now.Format(time.RFC3339),
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
			},
			Token: &dtos.TokenResponse{
				AccessToken:  "access_token_123",
				TokenType:    "Bearer",
				ExpiresAt:    now.Add(time.Hour).Format(time.RFC3339),
				RefreshToken: "refresh_token_123",
			},
			Session: &dtos.SessionResponse{
				ID:                "session-123",
				UserID:            "user-123",
				AccessTokenJTI:    stringPtr("jti_123"),
				UserAgent:         &userAgent,
				IPAddress:         &ipAddress,
				Location:          &location,
				DeviceFingerprint: &deviceFingerprint,
				CreatedAt:         now.Format(time.RFC3339),
				LastUsedAt:        now.Format(time.RFC3339),
				ExpiresAt:         now.Add(DefaultTokenExpirationHours * time.Hour).Format(time.RFC3339),
				IsActive:          true,
				RevokedAt:         nil,
				RevokedReason:     nil,
			},
		},
		EmptyLoginResponse: &dtos.LoginResponse{
			User:    nil,
			Token:   nil,
			Session: nil,
		},
		ValidUserResponse: &dtos.UserResponse{
			ID:                  "user-123",
			Email:               "test@example.com",
			FirstName:           "John",
			LastName:            "Doe",
			Phone:               &phone,
			IsActive:            true,
			IsVerified:          true,
			LastLoginAt:         stringPtr(now.Format(time.RFC3339)),
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			PasswordChangedAt:   now.Format(time.RFC3339),
			MustChangePassword:  false,
			CreatedAt:           now.Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
			Roles: []dtos.RoleResponse{
				{
					ID:           "role-user",
					Name:         "user",
					Description:  stringPtr("Regular user role"),
					IsSystemRole: true,
					IsActive:     true,
					CreatedAt:    now.Format(time.RFC3339),
					UpdatedAt:    now.Format(time.RFC3339),
				},
			},
		},
		AdminUserResponse: &dtos.UserResponse{
			ID:                  "admin-123",
			Email:               "admin@example.com",
			FirstName:           "Admin",
			LastName:            "User",
			Phone:               nil,
			IsActive:            true,
			IsVerified:          true,
			LastLoginAt:         stringPtr(now.Format(time.RFC3339)),
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          true,
			PasswordChangedAt:   now.Format(time.RFC3339),
			MustChangePassword:  false,
			CreatedAt:           now.Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
			Roles: []dtos.RoleResponse{
				{
					ID:           "role-admin",
					Name:         "admin",
					Description:  stringPtr("Administrator role"),
					IsSystemRole: true,
					IsActive:     true,
					CreatedAt:    now.Format(time.RFC3339),
					UpdatedAt:    now.Format(time.RFC3339),
				},
			},
		},
		EmptyUserResponse: &dtos.UserResponse{
			ID:                  "",
			Email:               "",
			FirstName:           "",
			LastName:            "",
			Phone:               nil,
			IsActive:            false,
			IsVerified:          false,
			LastLoginAt:         nil,
			FailedLoginAttempts: 0,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			PasswordChangedAt:   "",
			MustChangePassword:  false,
			CreatedAt:           "",
			UpdatedAt:           "",
			Roles:               []dtos.RoleResponse{},
		},
		InvalidUserResponse: &dtos.UserResponse{
			ID:                  "invalid-id",
			Email:               "invalid-email",
			FirstName:           "",
			LastName:            "",
			Phone:               nil,
			IsActive:            false,
			IsVerified:          false,
			LastLoginAt:         nil,
			FailedLoginAttempts: -1,
			LastFailedLoginAt:   nil,
			LockedUntil:         nil,
			MFAEnabled:          false,
			PasswordChangedAt:   "",
			MustChangePassword:  false,
			CreatedAt:           "",
			UpdatedAt:           "",
			Roles:               []dtos.RoleResponse{},
		},
		ValidRoleResponse: &dtos.RoleResponse{
			ID:           "role-user",
			Name:         "user",
			Description:  stringPtr("Regular user role"),
			IsSystemRole: true,
			IsActive:     true,
			CreatedAt:    now.Format(time.RFC3339),
			UpdatedAt:    now.Format(time.RFC3339),
		},
		AdminRoleResponse: &dtos.RoleResponse{
			ID:           "role-admin",
			Name:         "admin",
			Description:  stringPtr("Administrator role"),
			IsSystemRole: true,
			IsActive:     true,
			CreatedAt:    now.Format(time.RFC3339),
			UpdatedAt:    now.Format(time.RFC3339),
		},
		EmptyRoleResponse: &dtos.RoleResponse{
			ID:           "",
			Name:         "",
			Description:  nil,
			IsSystemRole: false,
			IsActive:     false,
			CreatedAt:    "",
			UpdatedAt:    "",
		},
		ValidSessionResponse: &dtos.SessionResponse{
			ID:                "session-123",
			UserID:            "user-123",
			AccessTokenJTI:    stringPtr("jti_123"),
			UserAgent:         &userAgent,
			IPAddress:         &ipAddress,
			Location:          &location,
			DeviceFingerprint: &deviceFingerprint,
			CreatedAt:         now.Format(time.RFC3339),
			LastUsedAt:        now.Format(time.RFC3339),
			ExpiresAt:         now.Add(DefaultTokenExpirationHours * time.Hour).Format(time.RFC3339),
			IsActive:          true,
			RevokedAt:         nil,
			RevokedReason:     nil,
		},
		ExpiredSessionResponse: &dtos.SessionResponse{
			ID:                "session-expired",
			UserID:            "user-123",
			AccessTokenJTI:    stringPtr("jti_expired"),
			UserAgent:         &userAgent,
			IPAddress:         &ipAddress,
			Location:          &location,
			DeviceFingerprint: &deviceFingerprint,
			CreatedAt:         now.Add(-DefaultTokenExpirationHours * time.Hour).Format(time.RFC3339),
			LastUsedAt:        now.Add(-1 * time.Hour).Format(time.RFC3339),
			ExpiresAt:         now.Add(-1 * time.Hour).Format(time.RFC3339),
			IsActive:          false,
			RevokedAt:         nil,
			RevokedReason:     nil,
		},
		EmptySessionResponse: &dtos.SessionResponse{
			ID:                "",
			UserID:            "",
			AccessTokenJTI:    nil,
			UserAgent:         nil,
			IPAddress:         nil,
			Location:          nil,
			DeviceFingerprint: nil,
			CreatedAt:         "",
			LastUsedAt:        "",
			ExpiresAt:         "",
			IsActive:          false,
			RevokedAt:         nil,
			RevokedReason:     nil,
		},
		ValidTokenResponse: &dtos.TokenResponse{
			AccessToken:  "access_token_123",
			TokenType:    "Bearer",
			ExpiresAt:    now.Add(time.Hour).Format(time.RFC3339),
			RefreshToken: "refresh_token_123",
		},
		ExpiredTokenResponse: &dtos.TokenResponse{
			AccessToken:  "expired_access_token_123",
			TokenType:    "Bearer",
			ExpiresAt:    now.Add(-time.Hour).Format(time.RFC3339),
			RefreshToken: "expired_refresh_token_123",
		},
		EmptyTokenResponse: &dtos.TokenResponse{
			AccessToken:  "",
			TokenType:    "",
			ExpiresAt:    "",
			RefreshToken: "",
		},
		ValidAuditLogResponse: &dtos.AuditLogResponse{
			ID:         "audit-123",
			UserID:     stringPtr("user-123"),
			SessionID:  stringPtr("session-123"),
			Action:     "user.login",
			Resource:   stringPtr("auth"),
			ResourceID: stringPtr("user-123"),
			Result:     domain.AuditResultSuccess,
			Message:    stringPtr("User logged in successfully"),
			IPAddress:  &ipAddress,
			UserAgent:  &userAgent,
			RequestID:  stringPtr("request-123"),
			CreatedAt:  now.Format(time.RFC3339),
		},
		FailedAuditLogResponse: &dtos.AuditLogResponse{
			ID:         "audit-failed-123",
			UserID:     nil,
			SessionID:  nil,
			Action:     "user.login",
			Resource:   stringPtr("auth"),
			ResourceID: nil,
			Result:     domain.AuditResultFailure,
			Message:    stringPtr("Invalid credentials"),
			IPAddress:  &ipAddress,
			UserAgent:  &userAgent,
			RequestID:  stringPtr("request-123"),
			CreatedAt:  now.Format(time.RFC3339),
		},
		EmptyAuditLogResponse: &dtos.AuditLogResponse{
			ID:         "",
			UserID:     nil,
			SessionID:  nil,
			Action:     "",
			Resource:   nil,
			ResourceID: nil,
			Result:     "",
			Message:    nil,
			IPAddress:  nil,
			UserAgent:  nil,
			RequestID:  nil,
			CreatedAt:  "",
		},
	}
}

// newAuthTestUseCases creates test use cases for the auth module
func newAuthTestUseCases() AuthTestUseCases {
	// GetUsersUseCase was moved to the users module
	// User-related use cases are now in the users module
	return AuthTestUseCases{}
}

// newAuthTestScenarios creates test scenarios for the auth module
func newAuthTestScenarios() AuthTestScenarios {
	now := time.Now()
	phone := "+1234567890"
	userAgent := DefaultUserAgent
	ipAddress := DefaultIPAddress

	return AuthTestScenarios{
		SuccessfulLogin: AuthLoginScenario{
			Request: &dtos.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			User: &entities.User{
				ID:                  "user-123",
				Email:               "test@example.com",
				PasswordHash:        "hashed_password_123",
				FirstName:           "John",
				LastName:            "Doe",
				Phone:               &phone,
				IsActive:            true,
				IsVerified:          true,
				LastLoginAt:         &now,
				FailedLoginAttempts: 0,
				LastFailedLoginAt:   nil,
				LockedUntil:         nil,
				MFAEnabled:          false,
				MFASecret:           nil,
				PasswordChangedAt:   now,
				MustChangePassword:  false,
				CreatedAt:           now,
				UpdatedAt:           now,
				Roles: []entities.Role{
					{
						ID:           "role-user",
						Name:         "user",
						Description:  stringPtr("Regular user role"),
						IsSystemRole: true,
						IsActive:     true,
						CreatedAt:    now,
						UpdatedAt:    now,
					},
				},
			},
			Token: &entities.AuthToken{
				AccessToken:  "access_token_123",
				TokenType:    "Bearer",
				ExpiresAt:    now.Add(time.Hour),
				RefreshToken: "refresh_token_123",
			},
			Session: &entities.Session{
				ID:                "session-123",
				UserID:            "user-123",
				RefreshTokenHash:  "hashed_refresh_token_123",
				AccessTokenJTI:    stringPtr("jti_123"),
				UserAgent:         &userAgent,
				IPAddress:         &ipAddress,
				Location:          nil,
				DeviceFingerprint: nil,
				CreatedAt:         now,
				LastUsedAt:        now,
				ExpiresAt:         now.Add(DefaultTokenExpirationHours * time.Hour),
				IsActive:          true,
				RevokedAt:         nil,
				RevokedReason:     nil,
			},
			ExpectedResponse: &dtos.LoginResponse{
				User: &dtos.UserResponse{
					ID:                  "user-123",
					Email:               "test@example.com",
					FirstName:           "John",
					LastName:            "Doe",
					Phone:               &phone,
					IsActive:            true,
					IsVerified:          true,
					LastLoginAt:         stringPtr(now.Format(time.RFC3339)),
					FailedLoginAttempts: 0,
					LastFailedLoginAt:   nil,
					LockedUntil:         nil,
					MFAEnabled:          false,
					PasswordChangedAt:   now.Format(time.RFC3339),
					MustChangePassword:  false,
					CreatedAt:           now.Format(time.RFC3339),
					UpdatedAt:           now.Format(time.RFC3339),
					Roles: []dtos.RoleResponse{
						{
							ID:           "role-user",
							Name:         "user",
							Description:  stringPtr("Regular user role"),
							IsSystemRole: true,
							IsActive:     true,
							CreatedAt:    now.Format(time.RFC3339),
							UpdatedAt:    now.Format(time.RFC3339),
						},
					},
				},
				Token: &dtos.TokenResponse{
					AccessToken:  "access_token_123",
					TokenType:    "Bearer",
					ExpiresAt:    now.Add(time.Hour).Format(time.RFC3339),
					RefreshToken: "refresh_token_123",
				},
				Session: &dtos.SessionResponse{
					ID:                "session-123",
					UserID:            "user-123",
					AccessTokenJTI:    stringPtr("jti_123"),
					UserAgent:         &userAgent,
					IPAddress:         &ipAddress,
					Location:          nil,
					DeviceFingerprint: nil,
					CreatedAt:         now.Format(time.RFC3339),
					LastUsedAt:        now.Format(time.RFC3339),
					ExpiresAt:         now.Add(DefaultTokenExpirationHours * time.Hour).Format(time.RFC3339),
					IsActive:          true,
					RevokedAt:         nil,
					RevokedReason:     nil,
				},
			},
			ExpectedError: "",
			UserAgent:     userAgent,
			IPAddress:     ipAddress,
		},
		FailedLogin: AuthLoginScenario{
			Request: &dtos.LoginRequest{
				Email:    "wrong@example.com",
				Password: "wrongpassword",
			},
			User:             nil,
			Token:            nil,
			Session:          nil,
			ExpectedResponse: nil,
			ExpectedError:    "invalid credentials",
			UserAgent:        userAgent,
			IPAddress:        ipAddress,
		},
		AccountLocked: AuthLoginScenario{
			Request: &dtos.LoginRequest{
				Email:    "locked@example.com",
				Password: "password123",
			},
			User: &entities.User{
				ID:                  "locked-user-123",
				Email:               "locked@example.com",
				PasswordHash:        "hashed_locked_password",
				FirstName:           "Locked",
				LastName:            "User",
				Phone:               nil,
				IsActive:            true,
				IsVerified:          true,
				LastLoginAt:         nil,
				FailedLoginAttempts: MaxFailedLoginAttempts,
				LastFailedLoginAt:   &now,
				LockedUntil:         timePtr(now.Add(time.Hour)),
				MFAEnabled:          false,
				MFASecret:           nil,
				PasswordChangedAt:   now,
				MustChangePassword:  false,
				CreatedAt:           now,
				UpdatedAt:           now,
				Roles:               []entities.Role{},
			},
			Token:            nil,
			Session:          nil,
			ExpectedResponse: nil,
			ExpectedError:    "account is locked",
			UserAgent:        userAgent,
			IPAddress:        ipAddress,
		},
		AccountInactive: AuthLoginScenario{
			Request: &dtos.LoginRequest{
				Email:    "inactive@example.com",
				Password: "password123",
			},
			User: &entities.User{
				ID:                  "inactive-user-123",
				Email:               "inactive@example.com",
				PasswordHash:        "hashed_inactive_password",
				FirstName:           "Inactive",
				LastName:            "User",
				Phone:               nil,
				IsActive:            false,
				IsVerified:          true,
				LastLoginAt:         nil,
				FailedLoginAttempts: 0,
				LastFailedLoginAt:   nil,
				LockedUntil:         nil,
				MFAEnabled:          false,
				MFASecret:           nil,
				PasswordChangedAt:   now,
				MustChangePassword:  false,
				CreatedAt:           now,
				UpdatedAt:           now,
				Roles:               []entities.Role{},
			},
			Token:            nil,
			Session:          nil,
			ExpectedResponse: nil,
			ExpectedError:    "account is not active",
			UserAgent:        userAgent,
			IPAddress:        ipAddress,
		},
		InvalidCredentials: AuthLoginScenario{
			Request: &dtos.LoginRequest{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			User: &entities.User{
				ID:                  "user-123",
				Email:               "user@example.com",
				PasswordHash:        "hashed_password_123",
				FirstName:           "John",
				LastName:            "Doe",
				Phone:               nil,
				IsActive:            true,
				IsVerified:          true,
				LastLoginAt:         nil,
				FailedLoginAttempts: MinFailedLoginAttempts,
				LastFailedLoginAt:   &now,
				LockedUntil:         nil,
				MFAEnabled:          false,
				MFASecret:           nil,
				PasswordChangedAt:   now,
				MustChangePassword:  false,
				CreatedAt:           now,
				UpdatedAt:           now,
				Roles:               []entities.Role{},
			},
			Token:            nil,
			Session:          nil,
			ExpectedResponse: nil,
			ExpectedError:    "invalid credentials",
			UserAgent:        userAgent,
			IPAddress:        ipAddress,
		},
		SuccessfulGetUsers: AuthGetUsersScenario{},
		FailedGetUsers:     AuthGetUsersScenario{},
		EmptyUsers:         AuthGetUsersScenario{},
		PaginationError:    AuthGetUsersScenario{},
	}
}

// GetTestEntity returns a specific test entity
func (s *AuthStubs) GetTestEntity(entityType string) any {
	switch entityType {
	case "valid_user":
		return s.Entities.ValidUser
	case "admin_user":
		return s.Entities.AdminUser
	case "inactive_user":
		return s.Entities.InactiveUser
	case "locked_user":
		return s.Entities.LockedUser
	case "unverified_user":
		return s.Entities.UnverifiedUser
	case "user_with_mfa":
		return s.Entities.UserWithMFA
	case "user_with_phone":
		return s.Entities.UserWithPhone
	case "empty_user":
		return s.Entities.EmptyUser
	case "invalid_user":
		return s.Entities.InvalidUser
	case "valid_role":
		return s.Entities.ValidRole
	case "admin_role":
		return s.Entities.AdminRole
	case "user_role":
		return s.Entities.UserRole
	case "guest_role":
		return s.Entities.GuestRole
	case "empty_role":
		return s.Entities.EmptyRole
	case "invalid_role":
		return s.Entities.InvalidRole
	case "valid_session":
		return s.Entities.ValidSession
	case "expired_session":
		return s.Entities.ExpiredSession
	case "revoked_session":
		return s.Entities.RevokedSession
	case "empty_session":
		return s.Entities.EmptySession
	case "invalid_session":
		return s.Entities.InvalidSession
	case "valid_audit_log":
		return s.Entities.ValidAuditLog
	case "failed_audit_log":
		return s.Entities.FailedAuditLog
	case "empty_audit_log":
		return s.Entities.EmptyAuditLog
	case "invalid_audit_log":
		return s.Entities.InvalidAuditLog
	case "valid_auth_token":
		return s.Entities.ValidAuthToken
	case "expired_auth_token":
		return s.Entities.ExpiredAuthToken
	case "empty_auth_token":
		return s.Entities.EmptyAuthToken
	case "invalid_auth_token":
		return s.Entities.InvalidAuthToken
	default:
		return s.Entities.ValidUser
	}
}

// GetTestDTO returns a specific test DTO
func (s *AuthStubs) GetTestDTO(dtoType string) any {
	switch dtoType {
	case "valid_login_request":
		return s.DTOs.ValidLoginRequest
	case "invalid_login_request":
		return s.DTOs.InvalidLoginRequest
	case "empty_login_request":
		return s.DTOs.EmptyLoginRequest
	case "valid_login_response":
		return s.DTOs.ValidLoginResponse
	case "empty_login_response":
		return s.DTOs.EmptyLoginResponse
	case "valid_user_response":
		return s.DTOs.ValidUserResponse
	case "admin_user_response":
		return s.DTOs.AdminUserResponse
	case "empty_user_response":
		return s.DTOs.EmptyUserResponse
	case "invalid_user_response":
		return s.DTOs.InvalidUserResponse
	case "valid_role_response":
		return s.DTOs.ValidRoleResponse
	case "admin_role_response":
		return s.DTOs.AdminRoleResponse
	case "empty_role_response":
		return s.DTOs.EmptyRoleResponse
	case "valid_session_response":
		return s.DTOs.ValidSessionResponse
	case "expired_session_response":
		return s.DTOs.ExpiredSessionResponse
	case "empty_session_response":
		return s.DTOs.EmptySessionResponse
	case "valid_token_response":
		return s.DTOs.ValidTokenResponse
	case "expired_token_response":
		return s.DTOs.ExpiredTokenResponse
	case "empty_token_response":
		return s.DTOs.EmptyTokenResponse
	case "valid_audit_log_response":
		return s.DTOs.ValidAuditLogResponse
	case "failed_audit_log_response":
		return s.DTOs.FailedAuditLogResponse
	case "empty_audit_log_response":
		return s.DTOs.EmptyAuditLogResponse
	default:
		return s.DTOs.ValidLoginRequest
	}
}

// GetTestUseCase returns a specific test use case
// GetUsersUseCase was moved to the users module
func (s *AuthStubs) GetTestUseCase(_ string) any {
	return nil
}

// GetTestScenario returns a specific test scenario
func (s *AuthStubs) GetTestScenario(scenarioType string) any {
	switch scenarioType {
	case "successful_login":
		return s.Scenarios.SuccessfulLogin
	case "failed_login":
		return s.Scenarios.FailedLogin
	case "account_locked":
		return s.Scenarios.AccountLocked
	case "account_inactive":
		return s.Scenarios.AccountInactive
	case "invalid_credentials":
		return s.Scenarios.InvalidCredentials
	case "successful_get_users":
		return s.Scenarios.SuccessfulGetUsers
	case "failed_get_users":
		return s.Scenarios.FailedGetUsers
	case "empty_users":
		return s.Scenarios.EmptyUsers
	case "pagination_error":
		return s.Scenarios.PaginationError
	default:
		return s.Scenarios.SuccessfulLogin
	}
}

// GetTestUser returns a test user by type (compatible with shared/test_utils)
func (s *AuthStubs) GetTestUser(userType string) *entities.User {
	now := time.Now()
	switch userType {
	case "admin":
		return &entities.User{
			ID: "admin-123", Email: "admin@example.com", PasswordHash: "hashed-admin-password",
			FirstName: "Admin", LastName: "User", IsActive: true, IsVerified: true,
			PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now,
			Roles: []entities.Role{
				{
					ID: "role-admin", Name: "admin", Description: stringPtr("Administrator role"),
					IsSystemRole: true, IsActive: true, CreatedAt: now, UpdatedAt: now,
				},
			},
		}
	case "regular":
		return &entities.User{
			ID: "user-456", Email: "user@example.com", PasswordHash: "hashed-user-password",
			FirstName: "Regular", LastName: "User", IsActive: true, IsVerified: true,
			PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now,
			Roles: []entities.Role{
				{
					ID: "role-user", Name: "user", Description: stringPtr("Regular user role"),
					IsSystemRole: true, IsActive: true, CreatedAt: now, UpdatedAt: now,
				},
			},
		}
	case "multi":
		return &entities.User{
			ID: "multi-789", Email: "multi@example.com", PasswordHash: "hashed-multi-password",
			FirstName: "Multi", LastName: "User", IsActive: true, IsVerified: true,
			PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now,
			Roles: []entities.Role{
				{
					ID: "role-user", Name: "user", Description: stringPtr("Regular user role"),
					IsSystemRole: true, IsActive: true, CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "role-admin", Name: "admin", Description: stringPtr("Administrator role"),
					IsSystemRole: true, IsActive: true, CreatedAt: now, UpdatedAt: now,
				},
			},
		}
	case "empty":
		return &entities.User{
			ID: "empty-999", Email: "empty@example.com", PasswordHash: "hashed-empty-password",
			FirstName: "Empty", LastName: "User", IsActive: true, IsVerified: true,
			PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now,
			Roles: []entities.Role{},
		}
	case "invalid":
		return &entities.User{
			ID: "", Email: "", PasswordHash: "", FirstName: "", LastName: "",
			IsActive: false, IsVerified: false, Roles: nil,
			PasswordChangedAt: time.Time{}, CreatedAt: time.Time{}, UpdatedAt: time.Time{},
		}
	default:
		return s.Entities.ValidUser
	}
}

// GetTestRole returns a test role by type (compatible with shared/test_utils)
func (s *AuthStubs) GetTestRole(roleType string) entities.Role {
	now := time.Now()
	switch roleType {
	case "admin":
		return entities.Role{
			ID: "role-admin", Name: "admin",
			Description:  stringPtr("Administrator role with full access"),
			IsSystemRole: true, IsActive: true, CreatedAt: now, UpdatedAt: now,
		}
	case "user":
		return entities.Role{
			ID: "role-user", Name: "user",
			Description:  stringPtr("Regular user role with limited access"),
			IsSystemRole: true, IsActive: true, CreatedAt: now, UpdatedAt: now,
		}
	case "guest":
		return entities.Role{
			ID: "role-guest", Name: "guest",
			Description:  stringPtr("Guest role with read-only access"),
			IsSystemRole: true, IsActive: true, CreatedAt: now, UpdatedAt: now,
		}
	case "empty":
		return entities.Role{}
	default:
		return s.Entities.UserRole
	}
}

// GetTestPassword returns a test password by type (compatible with shared/test_utils)
func (s *AuthStubs) GetTestPassword(passwordType string) string {
	switch passwordType {
	case "valid":
		return "mySecurePassword123"
	case "complex":
		return "P@ssw0rd!@#$%^&*()"
	case "empty":
		return ""
	case "long":
		return "veryLongPasswordWithManyCharactersToTestTheHashingAlgorithmAndEnsureItWorksCorrectlyWithLongInputs"
	case "special":
		return "!@#$%^&*()"
	case "unicode":
		return "pásswórd123"
	default:
		return "mySecurePassword123"
	}
}

// GetTestUsers returns a list of test users
func (s *AuthStubs) GetTestUsers(count int) []entities.User {
	users := []entities.User{}

	if count >= MinUserCount {
		users = append(users, *CloneUser(s.Entities.ValidUser))
	}

	if count >= MidUserCount {
		users = append(users, *CloneUser(s.Entities.AdminUser))
	}

	if count >= HighUserCount {
		users = append(users, *CloneUser(s.Entities.UserWithPhone))
	}

	if count >= MaxUserCountWithMFA {
		users = append(users, *CloneUser(s.Entities.UserWithMFA))
	}

	if count >= MaxUserCount {
		users = append(users, *CloneUser(s.Entities.InactiveUser))
	}

	for i := len(users); i < count; i++ {
		u := CloneUser(s.Entities.ValidUser)
		u.ID = uuid.New().String()
		u.Email = fmt.Sprintf("user%d@example.com", i+1)
		users = append(users, *u)
	}

	return users
}

// GetTestRoles returns a list of test roles
func (s *AuthStubs) GetTestRoles(count int) []entities.Role {
	roles := []entities.Role{}

	if count >= MinRoleCount {
		roles = append(roles, CloneRole(&s.Entities.ValidRole))
	}

	if count >= MidRoleCount {
		roles = append(roles, CloneRole(&s.Entities.AdminRole))
	}

	if count >= HighRoleCount {
		roles = append(roles, CloneRole(&s.Entities.UserRole))
	}

	if count >= MaxRoleCount {
		roles = append(roles, CloneRole(&s.Entities.GuestRole))
	}

	for i := len(roles); i < count; i++ {
		r := CloneRole(&s.Entities.ValidRole)
		r.ID = fmt.Sprintf("role-%d", i+1)
		r.Name = fmt.Sprintf("role%d", i+1)
		roles = append(roles, r)
	}

	return roles
}

// CreateMockUser creates a mock user with custom data
func (s *AuthStubs) CreateMockUser(email, firstName, lastName string, isActive, isVerified bool) *entities.User {
	now := time.Now()
	userID := uuid.New().String()

	return &entities.User{
		ID:                  userID,
		Email:               email,
		PasswordHash:        "hashed_password",
		FirstName:           firstName,
		LastName:            lastName,
		Phone:               nil,
		IsActive:            isActive,
		IsVerified:          isVerified,
		LastLoginAt:         nil,
		FailedLoginAttempts: 0,
		LastFailedLoginAt:   nil,
		LockedUntil:         nil,
		MFAEnabled:          false,
		MFASecret:           nil,
		PasswordChangedAt:   now,
		MustChangePassword:  false,
		CreatedAt:           now,
		UpdatedAt:           now,
		Roles:               []entities.Role{},
	}
}

// CreateMockLoginRequest creates a mock login request
func (s *AuthStubs) CreateMockLoginRequest(email, password string) *dtos.LoginRequest {
	return &dtos.LoginRequest{
		Email:    email,
		Password: password,
	}
}

// CreateMockGetUsersRequest creates a mock get users request
// GetUsersUseCase was moved to the users module
func (s *AuthStubs) CreateMockGetUsersRequest(limit, offset int) any {
	return nil
}
