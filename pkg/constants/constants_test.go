package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardHeaders(t *testing.T) {
	assert.Equal(t, "X-Request-ID", RequestIDHeader)
	assert.Equal(t, "Authorization", AuthorizationHeader)
	assert.Equal(t, "application/json", ContentTypeJSON)
	assert.Equal(t, "Bearer", BearerTokenType)
	assert.Equal(t, "Bearer ", BearerPrefix)
}

func TestPaginationLimits(t *testing.T) {
	assert.Equal(t, 20, DefaultPageSize)
	assert.Equal(t, 100, MaxPageSize)
	assert.Positive(t, MaxRequestSizeMB)
}

func TestTokenExpirationConstants(t *testing.T) {
	assert.Equal(t, 24, AccessTokenExpirationHours)
	assert.Equal(t, 7, RefreshTokenExpirationDays)
}

func TestRoleConstants(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin)
	assert.Equal(t, "viewer", RoleViewer)
}

func TestResponseCodeConstants(t *testing.T) {
	assert.Equal(t, "SUCCESS", SuccessCode)
	assert.Equal(t, "VALIDATION_ERROR", ValidationError)
	assert.Equal(t, "INTERNAL_ERROR", InternalError)
}

func TestCoreUserFacingMessages(t *testing.T) {
	assert.NotEmpty(t, SuccessMessage)
	assert.NotEmpty(t, CreatedMessage)
	assert.NotEmpty(t, InternalErrorMessage)
	assert.NotEmpty(t, GenericErrorMessage)
}

func TestAuthLoginMessages(t *testing.T) {
	assert.Equal(t, "Invalid credentials", InvalidCredentialsMessage)
	assert.NotEmpty(t, LoginEmailPasswordRequiredMessage)
	assert.NotEmpty(t, LoginSuccessMessage)
}

func TestReadinessMessages(t *testing.T) {
	assert.Contains(t, ServiceReadyToReceiveTrafficMessage, "ready")
	assert.Contains(t, ServiceNotReadyToReceiveTrafficMessage, "not ready")
}

func TestHealthStatusSlugs(t *testing.T) {
	assert.Equal(t, "healthy", HealthStatusHealthy)
	assert.Equal(t, "unhealthy", HealthStatusUnhealthy)
}

func TestPermissionConstants(t *testing.T) {
	assert.Equal(t, "read", PermissionRead)
	assert.Equal(t, "admin", PermissionAdmin)
}

func TestDateFormats(t *testing.T) {
	assert.Equal(t, "2006-01-02", DateFormat)
	assert.Contains(t, DateTimeFormat, "2006")
}

func TestStatusStrings(t *testing.T) {
	assert.Equal(t, "active", StatusActive)
	assert.Equal(t, "deleted", StatusDeleted)
}

func TestUploadLimits(t *testing.T) {
	assert.Equal(t, 10, MaxFilesPerUpload)
	assert.Equal(t, 50, MaxTotalSizeMB)
	assert.Equal(t, "application/pdf", AllowedPDFMimeTypes)
}
