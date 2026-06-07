package security

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// Constantes para strings duplicados en tests
const (
	TestSecretKey = "test-secret-key-32-bytes-long"
)

func tokenGenUserWithAdminRole() *entities.User {
	s := authtest.NewAuthStubs()
	u := authtest.CloneUser(s.GetTestUser("admin"))
	u.ID = "user-123"
	u.Email = "test@example.com"
	r := authtest.CloneRole(&s.Entities.AdminRole)
	r.ID = "role-1"
	u.Roles = []entities.Role{r}
	return u
}

func tokenGenUserNoRoles() *entities.User {
	s := authtest.NewAuthStubs()
	u := authtest.CloneUser(s.Entities.ValidUser)
	u.ID = "user-456"
	u.Email = "user@example.com"
	u.Roles = []entities.Role{}
	return u
}

func tokenGenUserMultiRole() *entities.User {
	s := authtest.NewAuthStubs()
	u := authtest.CloneUser(s.GetTestUser("multi"))
	u.ID = "user-789"
	u.Email = "multi@example.com"
	return u
}

func tokenGenConcurrentUser(i int) *entities.User {
	s := authtest.NewAuthStubs()
	u := authtest.CloneUser(s.Entities.ValidUser)
	u.ID = fmt.Sprintf("user-%d", i)
	u.Email = fmt.Sprintf("user%d@example.com", i)
	r := authtest.CloneRole(&s.Entities.ValidRole)
	r.ID = fmt.Sprintf("role-%d", i)
	r.Name = "user"
	u.Roles = []entities.Role{r}
	return u
}

func TestNewTokenGenerator(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "should create token generator with valid secret key",
			secretKey: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "should create token generator with empty secret key",
			secretKey: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := NewTokenGenerator(tt.secretKey)
			assert.NotNil(t, tg)
			assert.Equal(t, []byte(tt.secretKey), tg.secretKey)
		})
	}
}

func TestTokenGenerator_GenerateToken(t *testing.T) {
	secretKey := TestSecretKey
	tg := NewTokenGenerator(secretKey)

	tests := []struct {
		name        string
		user        *entities.User
		expectError bool
		validate    func(t *testing.T, authToken *entities.AuthToken, user *entities.User)
	}{
		{
			name:        "should generate valid token for user with role",
			user:        tokenGenUserWithAdminRole(),
			expectError: false,
			validate: func(t *testing.T, authToken *entities.AuthToken, user *entities.User) {
				assert.NotEmpty(t, authToken.AccessToken)
				assert.NotEmpty(t, authToken.RefreshToken)
				assert.Equal(t, constants.BearerTokenType, authToken.TokenType)
				assert.True(t, authToken.ExpiresAt.After(time.Now()))

				// Validate access token claims
				accessToken, err := jwt.Parse(authToken.AccessToken, func(token *jwt.Token) (any, error) {
					return []byte(secretKey), nil
				})
				require.NoError(t, err)
				require.True(t, accessToken.Valid)

				claims, ok := accessToken.Claims.(jwt.MapClaims)
				require.True(t, ok)
				assert.Equal(t, user.ID, claims[constants.JWTClaimUserID])
				assert.Equal(t, user.Email, claims[constants.JWTClaimEmail])
				assert.Equal(t, "admin", claims[constants.JWTClaimRole])
				assert.NotNil(t, claims[constants.JWTClaimExp])
				assert.NotNil(t, claims[constants.JWTClaimIat])
				assert.Empty(t, claims[constants.JWTClaimSessionID])

				// Validate refresh token claims
				refreshToken, err := jwt.Parse(authToken.RefreshToken, func(token *jwt.Token) (any, error) {
					return []byte(secretKey), nil
				})
				require.NoError(t, err)
				require.True(t, refreshToken.Valid)

				refreshClaims, ok := refreshToken.Claims.(jwt.MapClaims)
				require.True(t, ok)
				assert.Equal(t, user.ID, refreshClaims[constants.JWTClaimUserID])
				assert.Equal(t, constants.JWTTokenTypeRefresh, refreshClaims[constants.JWTClaimType])
				assert.NotNil(t, refreshClaims[constants.JWTClaimExp])
				assert.NotNil(t, refreshClaims[constants.JWTClaimIat])
				assert.Empty(t, refreshClaims[constants.JWTClaimSessionID])
			},
		},
		{
			name:        "should generate valid token for user without roles",
			user:        tokenGenUserNoRoles(),
			expectError: false,
			validate: func(t *testing.T, authToken *entities.AuthToken, user *entities.User) {
				assert.NotEmpty(t, authToken.AccessToken)
				assert.NotEmpty(t, authToken.RefreshToken)
				assert.Equal(t, constants.BearerTokenType, authToken.TokenType)

				// Validate access token claims
				accessToken, err := jwt.Parse(authToken.AccessToken, func(token *jwt.Token) (any, error) {
					return []byte(secretKey), nil
				})
				require.NoError(t, err)
				require.True(t, accessToken.Valid)

				claims, ok := accessToken.Claims.(jwt.MapClaims)
				require.True(t, ok)
				assert.Equal(t, user.ID, claims[constants.JWTClaimUserID])
				assert.Equal(t, user.Email, claims[constants.JWTClaimEmail])
				assert.Equal(t, "", claims[constants.JWTClaimRole]) // Empty string for user without roles
			},
		},
		{
			name:        "should generate valid token for user with multiple roles (uses first role)",
			user:        tokenGenUserMultiRole(),
			expectError: false,
			validate: func(t *testing.T, authToken *entities.AuthToken, user *entities.User) {
				accessToken, err := jwt.Parse(authToken.AccessToken, func(token *jwt.Token) (any, error) {
					return []byte(secretKey), nil
				})
				require.NoError(t, err)
				require.True(t, accessToken.Valid)

				claims, ok := accessToken.Claims.(jwt.MapClaims)
				require.True(t, ok)
				assert.Equal(t, "user", claims[constants.JWTClaimRole]) // Should use first role
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authToken, err := tg.GenerateToken(tt.user, "")

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, authToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, authToken)
				tt.validate(t, authToken, tt.user)
			}
		})
	}
}

func TestTokenGenerator_GenerateToken_ExpirationTimes(t *testing.T) {
	secretKey := TestSecretKey
	tg := NewTokenGenerator(secretKey)

	user := tokenGenUserWithAdminRole()

	authToken, err := tg.GenerateToken(user, "")
	require.NoError(t, err)
	require.NotNil(t, authToken)

	// Test access token expiration
	expectedAccessExpiration := time.Now().Add(time.Hour * constants.AccessTokenExpirationHours)
	assert.True(t, authToken.ExpiresAt.After(time.Now()))
	assert.True(t, authToken.ExpiresAt.Before(expectedAccessExpiration.Add(time.Minute))) // Allow 1 minute tolerance

	// Test access token JWT expiration
	accessToken, err := jwt.Parse(authToken.AccessToken, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	claims, ok := accessToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	exp, ok := claims[constants.JWTClaimExp].(float64)
	require.True(t, ok)

	expectedExp := time.Now().Add(time.Hour * constants.AccessTokenExpirationHours).Unix()
	assert.InDelta(t, expectedExp, int64(exp), 60) // Allow 60 seconds tolerance

	// Test refresh token JWT expiration
	refreshToken, err := jwt.Parse(authToken.RefreshToken, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	refreshClaims, ok := refreshToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	refreshExp, ok := refreshClaims[constants.JWTClaimExp].(float64)
	require.True(t, ok)

	expectedRefreshExp := time.Now().Add(time.Hour * 24 * constants.RefreshTokenExpirationDays).Unix()
	assert.InDelta(t, expectedRefreshExp, int64(refreshExp), 60) // Allow 60 seconds tolerance
}

func TestTokenGenerator_GenerateToken_InvalidSecretKey(t *testing.T) {
	// Test with empty secret key
	tg := NewTokenGenerator("")

	user := tokenGenUserWithAdminRole()

	authToken, err := tg.GenerateToken(user, "")
	// This should still work because JWT can be signed with an empty key
	assert.NoError(t, err)
	assert.NotNil(t, authToken)
}

func TestTokenGenerator_GenerateToken_ClaimsStructure(t *testing.T) {
	secretKey := TestSecretKey
	tg := NewTokenGenerator(secretKey)

	user := tokenGenUserWithAdminRole()

	authToken, err := tg.GenerateToken(user, "")
	require.NoError(t, err)

	// Parse and validate access token structure
	accessToken, err := jwt.Parse(authToken.AccessToken, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	claims, ok := accessToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Check all required claims exist
	requiredClaims := []string{
		constants.JWTClaimUserID,
		constants.JWTClaimEmail,
		constants.JWTClaimRole,
		constants.JWTClaimExp,
		constants.JWTClaimIat,
	}
	for _, claim := range requiredClaims {
		assert.Contains(t, claims, claim, "Access token missing required claim: %s", claim)
	}

	// Parse and validate refresh token structure
	refreshToken, err := jwt.Parse(authToken.RefreshToken, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	refreshClaims, ok := refreshToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Check all required claims exist for refresh token
	requiredRefreshClaims := []string{
		constants.JWTClaimUserID,
		constants.JWTClaimType,
		constants.JWTClaimExp,
		constants.JWTClaimIat,
	}
	for _, claim := range requiredRefreshClaims {
		assert.Contains(t, refreshClaims, claim, "Refresh token missing required claim: %s", claim)
	}
}

func TestTokenGenerator_GenerateToken_TokenValidation(t *testing.T) {
	secretKey := TestSecretKey
	tg := NewTokenGenerator(secretKey)

	user := tokenGenUserWithAdminRole()

	authToken, err := tg.GenerateToken(user, "")
	require.NoError(t, err)

	// Test that tokens can be validated with the correct secret
	accessToken, err := jwt.Parse(authToken.AccessToken, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)
	assert.True(t, accessToken.Valid)

	refreshToken, err := jwt.Parse(authToken.RefreshToken, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)
	assert.True(t, refreshToken.Valid)

	// Test that tokens cannot be validated with wrong secret
	wrongSecret := "wrong-secret-key"
	_, err = jwt.Parse(authToken.AccessToken, func(token *jwt.Token) (any, error) {
		return []byte(wrongSecret), nil
	})
	assert.Error(t, err)

	_, err = jwt.Parse(authToken.RefreshToken, func(token *jwt.Token) (any, error) {
		return []byte(wrongSecret), nil
	})
	assert.Error(t, err)
}

func TestTokenGenerator_GenerateToken_ConcurrentAccess(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long"
	tg := NewTokenGenerator(secretKey)

	// Test concurrent token generation with different users to ensure uniqueness
	const numGoroutines = 10
	results := make(chan *entities.AuthToken, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(userID int) {
			user := tokenGenConcurrentUser(userID)
			authToken, err := tg.GenerateToken(user, "")
			if err != nil {
				errors <- err
				return
			}
			results <- authToken
		}(i)
	}

	// Collect results
	tokens := make([]*entities.AuthToken, 0, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		select {
		case token := <-results:
			tokens = append(tokens, token)
		case err := <-errors:
			t.Fatalf("Unexpected error in concurrent token generation: %v", err)
		}
	}

	// Verify all tokens are valid and unique
	tokenStrings := make(map[string]bool)

	for _, token := range tokens {
		assert.NotEmpty(t, token.AccessToken)
		assert.NotEmpty(t, token.RefreshToken)
		assert.Equal(t, constants.BearerTokenType, token.TokenType)
		assert.True(t, token.ExpiresAt.After(time.Now()))

		// Check for uniqueness
		tokenKey := token.AccessToken + "|" + token.RefreshToken
		assert.False(t, tokenStrings[tokenKey], "Duplicate token generated")
		tokenStrings[tokenKey] = true
	}
}
