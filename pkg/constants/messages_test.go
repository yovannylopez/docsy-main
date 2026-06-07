package constants

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserNotFoundWithID(t *testing.T) {
	assert.Equal(t, "User with ID 'x' not found", UserNotFoundWithID("x"))
}

func TestUserNotFoundWithEmail(t *testing.T) {
	assert.Contains(t, UserNotFoundWithEmail("a@b.co"), "a@b.co")
}

func TestDocumentNotFoundWithID(t *testing.T) {
	assert.Contains(t, DocumentNotFoundWithID("d1"), "d1")
}

func TestFieldRequired(t *testing.T) {
	assert.Equal(t, "Field 'email' is required", FieldRequired("email"))
}

func TestFieldInvalid(t *testing.T) {
	got := FieldInvalid("name", "empty")
	assert.Contains(t, got, "name")
	assert.Contains(t, got, "empty")
}

func TestFieldTooLong(t *testing.T) {
	got := FieldTooLong("bio", 500)
	assert.Contains(t, got, "bio")
	assert.Contains(t, got, "500")
}

func TestFieldTooShort(t *testing.T) {
	got := FieldTooShort("pass", 8)
	assert.Contains(t, got, "pass")
	assert.Contains(t, got, "8")
}

func TestFileTooLarge(t *testing.T) {
	assert.Equal(t, "File 'x.pdf' exceeds the maximum size of 10 MB", FileTooLarge("x.pdf", 10))
}

func TestInvalidFileFormat(t *testing.T) {
	got := InvalidFileFormat("f.txt", []string{".pdf", ".png"})
	assert.Contains(t, got, "f.txt")
	assert.Contains(t, got, ".pdf")
}

func TestPermissionDeniedForResource(t *testing.T) {
	got := PermissionDeniedForResource("u1", "read", "reports")
	assert.Contains(t, got, "u1")
	assert.Contains(t, got, "read")
	assert.Contains(t, got, "reports")
}

func TestResourceCreatedWithID(t *testing.T) {
	got := ResourceCreatedWithID("User", "42")
	assert.Contains(t, got, "User")
	assert.Contains(t, got, "42")
}

func TestResourceUpdatedWithID(t *testing.T) {
	got := ResourceUpdatedWithID("Document", "99")
	assert.Contains(t, got, "updated")
	assert.Contains(t, got, "Document")
}

func TestResourceDeletedWithID(t *testing.T) {
	got := ResourceDeletedWithID("Series", "7")
	assert.Contains(t, got, "deleted")
	assert.Contains(t, got, "7")
}

func TestValidationFailedWithDetails(t *testing.T) {
	got := ValidationFailedWithDetails([]string{"a", "b"})
	assert.True(t, strings.HasPrefix(got, "Validation failed:"))
	assert.Contains(t, got, "a")
}

func TestRateLimitExceeded(t *testing.T) {
	got := RateLimitExceeded("10 req", "minute")
	assert.Contains(t, got, "10 req")
	assert.Contains(t, got, "minute")
}

func TestServiceUnavailable(t *testing.T) {
	got := ServiceUnavailable("email")
	assert.Contains(t, got, "email")
}

func TestDatabaseConnectionError(t *testing.T) {
	got := DatabaseConnectionError("timeout")
	assert.Contains(t, got, "timeout")
}

func TestExternalServiceError(t *testing.T) {
	got := ExternalServiceError("SMTP", "connection refused")
	assert.Contains(t, got, "SMTP")
	assert.Contains(t, got, "connection refused")
}
