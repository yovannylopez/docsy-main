package responses

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

func TestErrorType_Constants(t *testing.T) {
	// Verify all constants are defined correctly
	assert.Equal(t, ErrorType("VALIDATION_ERROR"), ValidationError)
	assert.Equal(t, ErrorType("AUTHENTICATION_ERROR"), AuthenticationError)
	assert.Equal(t, ErrorType("AUTHORIZATION_ERROR"), AuthorizationError)
	assert.Equal(t, ErrorType("NOT_FOUND_ERROR"), NotFoundError)
	assert.Equal(t, ErrorType("CONFLICT_ERROR"), ConflictError)
	assert.Equal(t, ErrorType("DUPLICATE_ERROR"), DuplicateError)
	assert.Equal(t, ErrorType("INTERNAL_SERVER_ERROR"), InternalServerError)
	assert.Equal(t, ErrorType("SERVICE_ERROR"), ServiceError)
	assert.Equal(t, ErrorType("DATABASE_ERROR"), DatabaseError)
	assert.Equal(t, ErrorType("BAD_REQUEST_ERROR"), BadRequestError)
	assert.Equal(t, ErrorType("UNPROCESSABLE_ERROR"), UnprocessableError)
}

func TestNewAppError(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		code      int
		message   string
		expected  *AppError
	}{
		{
			name:      "validation error",
			errorType: ValidationError,
			code:      http_status.BadRequest.Code,
			message:   "Required field",
			expected: &AppError{
				Type:    ValidationError,
				Code:    http_status.BadRequest.Code,
				Message: "Required field",
			},
		},
		{
			name:      "authentication error",
			errorType: AuthenticationError,
			code:      http_status.Unauthorized.Code,
			message:   "Invalid token",
			expected: &AppError{
				Type:    AuthenticationError,
				Code:    http_status.Unauthorized.Code,
				Message: "Invalid token",
			},
		},
		{
			name:      "internal error",
			errorType: InternalServerError,
			code:      http_status.InternalError.Code,
			message:   "Internal server error",
			expected: &AppError{
				Type:    InternalServerError,
				Code:    http_status.InternalError.Code,
				Message: "Internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			error := NewAppError(tt.errorType, tt.code, tt.message)

			assert.NotNil(t, error)
			assert.Equal(t, tt.expected.Type, error.Type)
			assert.Equal(t, tt.expected.Code, error.Code)
			assert.Equal(t, tt.expected.Message, error.Message)
			assert.Empty(t, error.Description)
			assert.Nil(t, error.Details)
			assert.Empty(t, error.RequestID)
		})
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "validation error",
			appError: &AppError{
				Type:    ValidationError,
				Code:    http_status.BadRequest.Code,
				Message: "Required field",
			},
			expected: "[VALIDATION_ERROR] Required field",
		},
		{
			name: "authentication error",
			appError: &AppError{
				Type:    AuthenticationError,
				Code:    http_status.Unauthorized.Code,
				Message: "Invalid token",
			},
			expected: "[AUTHENTICATION_ERROR] Invalid token",
		},
		{
			name: "internal error",
			appError: &AppError{
				Type:    InternalServerError,
				Code:    http_status.InternalError.Code,
				Message: "Internal server error",
			},
			expected: "[INTERNAL_SERVER_ERROR] Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_WithDescription(t *testing.T) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Required field")
	description := "The email field is required to continue"

	result := appError.WithDescription(description)

	// Verify it returns the same object (method chaining)
	assert.Equal(t, appError, result)

	// Verify the description was added
	assert.Equal(t, description, appError.Description)
}

func TestAppError_WithDetails(t *testing.T) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Invalid fields")
	details := map[string]any{
		"email":    "Invalid email format",
		"password": "Password must be at least 8 characters",
		"age":      "Age must be greater than 0",
	}

	result := appError.WithDetails(details)

	// Verify it returns the same object (method chaining)
	assert.Equal(t, appError, result)

	// Verify the details were added
	assert.Equal(t, details, appError.Details)
}

func TestAppError_WithRequestID(t *testing.T) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Required field")
	requestID := "req_123456789"

	result := appError.WithRequestID(requestID)

	// Verify it returns the same object (method chaining)
	assert.Equal(t, appError, result)

	// Verify the request ID was added
	assert.Equal(t, requestID, appError.RequestID)
}

func TestAppError_MethodChaining(t *testing.T) {
	// Test that methods can be chained
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Invalid fields").
		WithDescription("Fields have validation errors").
		WithDetails(map[string]any{
			"email": "Invalid format",
		}).
		WithRequestID("req_123")

	// Verify all fields were set correctly
	assert.Equal(t, ValidationError, appError.Type)
	assert.Equal(t, http_status.BadRequest.Code, appError.Code)
	assert.Equal(t, "Invalid fields", appError.Message)
	assert.Equal(t, "Fields have validation errors", appError.Description)
	assert.Equal(t, map[string]any{"email": "Invalid format"}, appError.Details)
	assert.Equal(t, "req_123", appError.RequestID)
}

func TestNewValidationError(t *testing.T) {
	message := "Email field is required"
	error := NewValidationError(message)

	assert.NotNil(t, error)
	assert.Equal(t, ValidationError, error.Type)
	assert.Equal(t, http_status.BadRequest.Code, error.Code)
	assert.Equal(t, message, error.Message)
}

func TestNewAuthenticationError(t *testing.T) {
	message := "Invalid authentication token"
	error := NewAuthenticationError(message)

	assert.NotNil(t, error)
	assert.Equal(t, AuthenticationError, error.Type)
	assert.Equal(t, http_status.Unauthorized.Code, error.Code)
	assert.Equal(t, message, error.Message)
}

func TestNewAuthorizationError(t *testing.T) {
	message := "You don't have permissions to access this resource"
	error := NewAuthorizationError(message)

	assert.NotNil(t, error)
	assert.Equal(t, AuthorizationError, error.Type)
	assert.Equal(t, http_status.Forbidden.Code, error.Code)
	assert.Equal(t, message, error.Message)
}

func TestNewNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		expected string
	}{
		{
			name:     "user resource",
			resource: "user",
			expected: "user: not found",
		},
		{
			name:     "product resource",
			resource: "product",
			expected: "product: not found",
		},
		{
			name:     "resource with spaces",
			resource: "user profile",
			expected: "user profile: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			error := NewNotFoundError(tt.resource)

			assert.NotNil(t, error)
			assert.Equal(t, NotFoundError, error.Type)
			assert.Equal(t, http_status.NotFound.Code, error.Code)
			assert.Equal(t, tt.expected, error.Message)
		})
	}
}

func TestNewConflictError(t *testing.T) {
	message := "The user already exists in the system"
	error := NewConflictError(message)

	assert.NotNil(t, error)
	assert.Equal(t, ConflictError, error.Type)
	assert.Equal(t, http_status.Conflict.Code, error.Code)
	assert.Equal(t, message, error.Message)
}

func TestNewDuplicateError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "email field",
			field:    "email",
			expected: "email is already in use",
		},
		{
			name:     "username field",
			field:    "username",
			expected: "username is already in use",
		},
		{
			name:     "field with spaces",
			field:    "phone number",
			expected: "phone number is already in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			error := NewDuplicateError(tt.field)

			assert.NotNil(t, error)
			assert.Equal(t, DuplicateError, error.Type)
			assert.Equal(t, http_status.Conflict.Code, error.Code)
			assert.Equal(t, tt.expected, error.Message)
		})
	}
}

func TestNewInternalError(t *testing.T) {
	message := "Internal server error"
	error := NewInternalError(message)

	assert.NotNil(t, error)
	assert.Equal(t, InternalServerError, error.Type)
	assert.Equal(t, http_status.InternalError.Code, error.Code)
	assert.Equal(t, message, error.Message)
}

func TestNewBadRequestError(t *testing.T) {
	message := "Invalid request"
	error := NewBadRequestError(message)

	assert.NotNil(t, error)
	assert.Equal(t, BadRequestError, error.Type)
	assert.Equal(t, http_status.BadRequest.Code, error.Code)
	assert.Equal(t, message, error.Message)
}

func TestAppError_JSONSerialization(t *testing.T) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Invalid fields").
		WithDescription("Fields have validation errors").
		WithDetails(map[string]any{
			"email":    "Invalid email format",
			"password": "Password must be at least 8 characters",
		}).
		WithRequestID("req_123456789")

	// Serialize to JSON
	jsonData, err := json.Marshal(appError)
	assert.NoError(t, err)

	// Deserialize from JSON
	var result AppError
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verify all fields were serialized correctly
	assert.Equal(t, appError.Type, result.Type)
	assert.Equal(t, appError.Code, result.Code)
	assert.Equal(t, appError.Message, result.Message)
	assert.Equal(t, appError.Description, result.Description)
	assert.Equal(t, appError.Details, result.Details)
	assert.Equal(t, appError.RequestID, result.RequestID)
}

func TestAppError_JSONSerialization_Minimal(t *testing.T) {
	// Test minimal error serialization (without optional fields)
	appError := NewAppError(InternalServerError, http_status.InternalError.Code, "Internal error")

	jsonData, err := json.Marshal(appError)
	assert.NoError(t, err)

	var result AppError
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verify optional fields are empty
	assert.Equal(t, appError.Type, result.Type)
	assert.Equal(t, appError.Code, result.Code)
	assert.Equal(t, appError.Message, result.Message)
	assert.Empty(t, result.Description)
	assert.Nil(t, result.Details)
	assert.Empty(t, result.RequestID)
}

func TestAppError_JSONStructure(t *testing.T) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Invalid fields").
		WithDescription("Fields have validation errors").
		WithDetails(map[string]any{
			"email": "Invalid format",
		}).
		WithRequestID("req_123")

	jsonData, err := json.Marshal(appError)
	assert.NoError(t, err)

	// Verify JSON structure
	var jsonMap map[string]any
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	// Verify all fields are present
	assert.Equal(t, "VALIDATION_ERROR", jsonMap["type"])
	assert.Equal(t, float64(400), jsonMap["code"])
	assert.Equal(t, "Invalid fields", jsonMap["message"])
	assert.Equal(t, "Fields have validation errors", jsonMap["description"])
	assert.Equal(t, "req_123", jsonMap["request_id"])

	// Verify details
	details, ok := jsonMap["details"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "Invalid format", details["email"])
}

func TestAppError_EmptyFields(t *testing.T) {
	// Test that optional fields are handled correctly when empty
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Error")

	jsonData, err := json.Marshal(appError)
	assert.NoError(t, err)

	var jsonMap map[string]any
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)

	// Verify optional fields are not present
	assert.Contains(t, jsonMap, "type")
	assert.Contains(t, jsonMap, "code")
	assert.Contains(t, jsonMap, "message")
	assert.NotContains(t, jsonMap, "description")
	assert.NotContains(t, jsonMap, "details")
	assert.NotContains(t, jsonMap, "request_id")
}

func TestAppError_ComplexDetails(t *testing.T) {
	// Test with complex details
	complexDetails := map[string]any{
		"errors": []string{
			"Invalid email",
			"Password too short",
		},
		"field_count": 2,
		"nested": map[string]any{
			"email": map[string]any{
				"value": "invalid-email",
				"rule":  "email_format",
			},
		},
	}

	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Multiple errors").
		WithDetails(complexDetails)

	jsonData, err := json.Marshal(appError)
	assert.NoError(t, err)

	var result AppError
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verify details were serialized correctly
	// Note: JSON marshaling changes types, so we verify the structure
	assert.NotNil(t, result.Details)

	// Verify it has the expected fields
	details := result.Details
	assert.NotNil(t, details)

	// Verify errors
	errors, ok := details["errors"].([]any)
	assert.True(t, ok)
	assert.Len(t, errors, 2)
	assert.Equal(t, "Invalid email", errors[0])
	assert.Equal(t, "Password too short", errors[1])

	// Verify field_count
	fieldCount, ok := details["field_count"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(2), fieldCount)

	// Verify nested
	nested, ok := details["nested"].(map[string]any)
	assert.True(t, ok)

	email, ok := nested["email"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "invalid-email", email["value"])
	assert.Equal(t, "email_format", email["rule"])
}

func TestAppError_ErrorInterface(t *testing.T) {
	// Verify AppError implements the error interface
	var _ error = &AppError{}

	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Test error")

	// Verify it can be used as an error
	var err error = appError
	assert.Equal(t, "[VALIDATION_ERROR] Test error", err.Error())
}

func BenchmarkNewAppError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAppError(ValidationError, http_status.BadRequest.Code, "Test error")
	}
}

func BenchmarkAppError_WithDescription(b *testing.B) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appError.WithDescription("Test description")
	}
}

func BenchmarkAppError_WithDetails(b *testing.B) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Test error")
	details := map[string]any{
		"field1": "error1",
		"field2": "error2",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appError.WithDetails(details)
	}
}

func BenchmarkAppError_WithRequestID(b *testing.B) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appError.WithRequestID("req_123456789")
	}
}

func BenchmarkAppError_Error(b *testing.B) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Test error message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appError.Error()
	}
}

func BenchmarkAppError_JSONMarshal(b *testing.B) {
	appError := NewAppError(ValidationError, http_status.BadRequest.Code, "Test error").
		WithDescription("Test description").
		WithDetails(map[string]any{
			"field1": "error1",
			"field2": "error2",
		}).
		WithRequestID("req_123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(appError)
	}
}

func BenchmarkNewValidationError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewValidationError("Test validation error")
	}
}

func BenchmarkNewNotFoundError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNotFoundError("user")
	}
}

func BenchmarkNewDuplicateError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewDuplicateError("email")
	}
}
