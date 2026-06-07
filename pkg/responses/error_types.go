package responses

import (
	"fmt"

	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

// ErrorType defines the error type
type ErrorType string

const (
	// ValidationError represents validation errors
	ValidationError ErrorType = "VALIDATION_ERROR"

	// AuthenticationError represents authentication errors
	AuthenticationError ErrorType = "AUTHENTICATION_ERROR"
	// AuthorizationError represents authorization errors
	AuthorizationError ErrorType = "AUTHORIZATION_ERROR"

	// NotFoundError represents resource not found errors
	NotFoundError ErrorType = "NOT_FOUND_ERROR"
	// ConflictError represents conflict errors
	ConflictError ErrorType = "CONFLICT_ERROR"
	// DuplicateError represents duplicate errors
	DuplicateError ErrorType = "DUPLICATE_ERROR"

	// InternalServerError represents internal system errors
	InternalServerError ErrorType = "INTERNAL_SERVER_ERROR"
	// ServiceError represents service errors
	ServiceError ErrorType = "SERVICE_ERROR"
	// DatabaseError represents database errors
	DatabaseError ErrorType = "DATABASE_ERROR"

	// BadRequestError represents invalid request errors
	BadRequestError ErrorType = "BAD_REQUEST_ERROR"
	// UnprocessableError represents unprocessable entity errors
	UnprocessableError ErrorType = "UNPROCESSABLE_ERROR"
)

// AppError represents an application error
type AppError struct {
	Type        ErrorType      `json:"type"`
	Code        int            `json:"code"`
	Message     string         `json:"message"`
	Description string         `json:"description,omitempty"`
	Details     map[string]any `json:"details,omitempty"`
	RequestID   string         `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code int, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
	}
}

// WithDescription adds a description to the error
func (e *AppError) WithDescription(description string) *AppError {
	e.Description = description
	return e
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details map[string]any) *AppError {
	e.Details = details
	return e
}

// WithRequestID adds a request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *AppError {
	return NewAppError(ValidationError, http_status.BadRequest.Code, message)
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *AppError {
	return NewAppError(AuthenticationError, http_status.Unauthorized.Code, message)
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string) *AppError {
	return NewAppError(AuthorizationError, http_status.Forbidden.Code, message)
}

// NewNotFoundError creates a new resource not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError(NotFoundError, http_status.NotFound.Code, fmt.Sprintf("%s: not found", resource))
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *AppError {
	return NewAppError(ConflictError, http_status.Conflict.Code, message)
}

// NewDuplicateError creates a new duplicate error
func NewDuplicateError(field string) *AppError {
	return NewAppError(DuplicateError, http_status.Conflict.Code, fmt.Sprintf("%s is already in use", field))
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *AppError {
	return NewAppError(InternalServerError, http_status.InternalError.Code, message)
}

// NewBadRequestError creates a new invalid request error
func NewBadRequestError(message string) *AppError {
	return NewAppError(BadRequestError, http_status.BadRequest.Code, message)
}
