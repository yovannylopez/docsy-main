package test_utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/security"
	sharedTestUtils "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

// TestUsingStubs combines stubs from shared (config) and auth (users/roles/passwords).
func TestUsingStubs(t *testing.T) {
	err := logging.Init(false)
	require.NoError(t, err)

	configStubs := sharedTestUtils.NewStubs()
	authStubs := NewAuthStubs()

	validConfig := configStubs.GetTestConfig("valid")
	invalidConfig := configStubs.GetTestConfig("invalid")
	devConfig := configStubs.GetTestConfig(constants.EnvDevelopment)
	assert.NotNil(t, validConfig)
	assert.NotNil(t, invalidConfig)
	assert.NotNil(t, devConfig)
	assert.Equal(t, "localhost", validConfig.Database.Host)
	assert.Equal(t, "invalid-host", invalidConfig.Database.Host)

	adminUser := authStubs.GetTestUser("admin")
	regularUser := authStubs.GetTestUser("regular")
	emptyUser := authStubs.GetTestUser("empty")
	assert.Equal(t, "admin-123", adminUser.ID)
	assert.Equal(t, "admin@example.com", adminUser.Email)
	assert.Equal(t, "hashed-admin-password", adminUser.PasswordHash)
	assert.True(t, adminUser.IsActive)
	assert.True(t, adminUser.IsVerified)
	assert.Len(t, adminUser.Roles, 1)
	assert.Equal(t, "admin", adminUser.Roles[0].Name)

	assert.Equal(t, "user-456", regularUser.ID)
	assert.Equal(t, "user@example.com", regularUser.Email)
	assert.Len(t, regularUser.Roles, 1)
	assert.Equal(t, "user", regularUser.Roles[0].Name)

	assert.Equal(t, "empty-999", emptyUser.ID)
	assert.Len(t, emptyUser.Roles, 0)

	validPassword := authStubs.GetTestPassword("valid")
	complexPassword := authStubs.GetTestPassword("complex")
	emptyPassword := authStubs.GetTestPassword("empty")
	assert.Equal(t, "mySecurePassword123", validPassword)
	assert.Equal(t, "P@ssw0rd!@#$%^&*()", complexPassword)
	assert.Equal(t, "", emptyPassword)

	adminRole := authStubs.GetTestRole("admin")
	userRole := authStubs.GetTestRole("user")
	guestRole := authStubs.GetTestRole("guest")
	assert.Equal(t, "role-admin", adminRole.ID)
	assert.Equal(t, "admin", adminRole.Name)
	assert.True(t, adminRole.IsSystemRole)
	assert.True(t, adminRole.IsActive)
	assert.Equal(t, "role-user", userRole.ID)
	assert.Equal(t, "user", userRole.Name)
	assert.Equal(t, "role-guest", guestRole.ID)
	assert.Equal(t, "guest", guestRole.Name)
}

// TestTokenGenerationWithStubs verifies JWT generation with auth test users.
func TestTokenGenerationWithStubs(t *testing.T) {
	err := logging.Init(false)
	require.NoError(t, err)

	authStubs := NewAuthStubs()
	tokenGenerator := security.NewTokenGenerator("test-secret-key-32-bytes-long")

	testCases := []struct {
		name        string
		userType    string
		expectError bool
	}{
		{name: "admin user with role", userType: "admin", expectError: false},
		{name: "regular user with role", userType: "regular", expectError: false},
		{name: "user with multiple roles", userType: "multi", expectError: false},
		{name: "user without roles", userType: "empty", expectError: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := authStubs.GetTestUser(tc.userType)
			authToken, err := tokenGenerator.GenerateToken(user, "")

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, authToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, authToken)
				assert.NotEmpty(t, authToken.AccessToken)
				assert.NotEmpty(t, authToken.RefreshToken)
				assert.Equal(t, "Bearer", authToken.TokenType)
				assert.True(t, authToken.ExpiresAt.After(time.Now()))
			}
		})
	}
}

// TestUserValidationWithStubs validates test user data.
func TestUserValidationWithStubs(t *testing.T) {
	err := logging.Init(false)
	require.NoError(t, err)

	authStubs := NewAuthStubs()

	testCases := []struct {
		name        string
		userType    string
		expectValid bool
	}{
		{name: "admin user", userType: "admin", expectValid: true},
		{name: "regular user", userType: "regular", expectValid: true},
		{name: "user with multiple roles", userType: "multi", expectValid: true},
		{name: "empty user", userType: "empty", expectValid: true},
		{name: "invalid user", userType: "invalid", expectValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := authStubs.GetTestUser(tc.userType)

			if tc.expectValid {
				assert.NotEmpty(t, user.ID)
				assert.NotEmpty(t, user.Email)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEmpty(t, user.FirstName)
				assert.NotEmpty(t, user.LastName)
				assert.True(t, user.IsActive)
				assert.True(t, user.IsVerified)
			} else {
				assert.Empty(t, user.ID)
				assert.Empty(t, user.Email)
				assert.Empty(t, user.PasswordHash)
				assert.Empty(t, user.FirstName)
				assert.Empty(t, user.LastName)
				assert.False(t, user.IsActive)
				assert.False(t, user.IsVerified)
			}
		})
	}
}
