package constants

import "fmt"

// UserNotFoundWithID generates a message of user not found with ID.
func UserNotFoundWithID(userID string) string {
	return fmt.Sprintf("User with ID '%s' not found", userID)
}

// UserNotFoundWithEmail generates a message of user not found with email.
func UserNotFoundWithEmail(email string) string {
	return fmt.Sprintf("User with email '%s' not found", email)
}

// DocumentNotFoundWithID generates a message of document not found with ID.
func DocumentNotFoundWithID(documentID string) string {
	return fmt.Sprintf("Document with ID '%s' not found", documentID)
}

// FieldRequired generates a message of field required.
func FieldRequired(fieldName string) string {
	return fmt.Sprintf("Field '%s' is required", fieldName)
}

// FieldInvalid generates a message of field invalid.
func FieldInvalid(fieldName, reason string) string {
	return fmt.Sprintf("Field '%s' is invalid: %s", fieldName, reason)
}

// FieldTooLong generates a message of field too long.
func FieldTooLong(fieldName string, maxLength int) string {
	return fmt.Sprintf("Field '%s' exceeds the maximum length of %d characters", fieldName, maxLength)
}

// FieldTooShort generates a message of field too short.
func FieldTooShort(fieldName string, minLength int) string {
	return fmt.Sprintf("Field '%s' must have at least %d characters", fieldName, minLength)
}

// FileTooLarge generates a message of file too large.
func FileTooLarge(fileName string, maxSizeMB int) string {
	return fmt.Sprintf("File '%s' exceeds the maximum size of %d MB", fileName, maxSizeMB)
}

// InvalidFileFormat generates a message of invalid file format.
func InvalidFileFormat(fileName string, allowedFormats []string) string {
	return fmt.Sprintf("File '%s' has an invalid format. Allowed formats: %v", fileName, allowedFormats)
}

// PermissionDeniedForResource generates a message of permission denied for a resource.
func PermissionDeniedForResource(userID, resource, action string) string {
	return fmt.Sprintf("User '%s' does not have permissions for %s in the resource '%s'", userID, action, resource)
}

// ResourceCreatedWithID generates a message of resource created with ID.
func ResourceCreatedWithID(resourceType, resourceID string) string {
	return fmt.Sprintf("%s created successfully with ID: %s", resourceType, resourceID)
}

// ResourceUpdatedWithID generates a message of resource updated with ID.
func ResourceUpdatedWithID(resourceType, resourceID string) string {
	return fmt.Sprintf("%s updated successfully with ID: %s", resourceType, resourceID)
}

// ResourceDeletedWithID generates a message of resource deleted with ID.
func ResourceDeletedWithID(resourceType, resourceID string) string {
	return fmt.Sprintf("%s deleted successfully with ID: %s", resourceType, resourceID)
}

// ValidationFailedWithDetails generates a message of validation failed with details.
func ValidationFailedWithDetails(details []string) string {
	return fmt.Sprintf("Validation failed: %v", details)
}

// RateLimitExceeded generates a message of rate limit exceeded.
func RateLimitExceeded(limit, window string) string {
	return fmt.Sprintf("Rate limit exceeded: maximum %s per %s", limit, window)
}

// ServiceUnavailable generates a message of service unavailable.
func ServiceUnavailable(serviceName string) string {
	return fmt.Sprintf("Service '%s' is not available at the moment", serviceName)
}

// DatabaseConnectionError generates a message of database connection error.
func DatabaseConnectionError(details string) string {
	return fmt.Sprintf("Database connection error: %s", details)
}

// ExternalServiceError generates a message of external service error.
func ExternalServiceError(serviceName, details string) string {
	return fmt.Sprintf("External service '%s' error: %s", serviceName, details)
}
