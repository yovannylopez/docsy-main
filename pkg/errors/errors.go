package errors

import (
	"fmt"

	pkgerrs "github.com/pkg/errors"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound represents not found errors
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeUnauthorized represents unauthorized errors
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	// ErrorTypeForbidden represents forbidden errors
	ErrorTypeForbidden ErrorType = "forbidden"
	// ErrorTypeConflict represents conflict errors
	ErrorTypeConflict ErrorType = "conflict"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeDatabase represents database errors
	ErrorTypeDatabase ErrorType = "database"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "external"
	// ErrorTypeServiceUnavailable represents service unavailable errors
	ErrorTypeServiceUnavailable ErrorType = "service_unavailable"
)

// AppError represents an application error with additional context
type AppError struct {
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Operation   string    `json:"operation,omitempty"`
	Resource    string    `json:"resource,omitempty"`
	UserMessage string    `json:"user_message,omitempty"`
	cause       error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}

	return e.Message
}

// Cause returns the original error
func (e *AppError) Cause() error {
	return e.cause
}

// Unwrap implements the errors.Wrapper interface
func (e *AppError) Unwrap() error {
	return e.cause
}

// Format implements fmt.Formatter
func (e *AppError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%s: %s", e.Type, e.Message)
			if e.cause != nil {
				_, _ = fmt.Fprintf(s, "\nCaused by: %+v", e.cause)
			}

			return
		}

		fallthrough
	case 's':
		_, _ = fmt.Fprintf(s, "%s: %s", e.Type, e.Message)
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

// New creates a new application error
func New(errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context.
// If err is *AppError, returns a new *AppError with the prefixed message and cause pointing to the
// original error (without mutating the received pointer).
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Type:        appErr.Type,
			Code:        appErr.Code,
			Message:     fmt.Sprintf("%s: %s", message, appErr.Message),
			Details:     appErr.Details,
			Operation:   appErr.Operation,
			Resource:    appErr.Resource,
			UserMessage: appErr.UserMessage,
			cause:       appErr,
		}
	}

	return pkgerrs.Wrap(err, message)
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...any) error {
	return pkgerrs.Wrapf(err, format, args...)
}

// WithStack adds a stack trace to the error
func WithStack(err error) error {
	return pkgerrs.WithStack(err)
}

// Cause returns the root error of the chain according to github.com/pkg/errors, without adding
// additional fmt.Errorf wrappers (preserves identity for errors.Is / errors.As at the root).
func Cause(err error) error {
	if err == nil {
		return nil
	}

	return pkgerrs.Cause(err)
}

// Is checks if the error is of a specific type
func Is(err error, target error) bool {
	return pkgerrs.Is(err, target)
}

// As checks if the error is of a specific type and extracts it
func As(err error, target any) bool {
	return pkgerrs.As(err, target)
}

// ValidationError creates a validation error
func ValidationError(code, message string) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
	}
}

// NotFoundError creates a not found error
func NotFoundError(resource, identifier string) *AppError {
	return &AppError{
		Type:     ErrorTypeNotFound,
		Code:     "RESOURCE_NOT_FOUND",
		Message:  fmt.Sprintf("%s not found: %s", resource, identifier),
		Resource: resource,
	}
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Code:    "FORBIDDEN",
		Message: message,
	}
}

// ConflictError creates a conflict error
func ConflictError(resource, identifier string) *AppError {
	return &AppError{
		Type:     ErrorTypeConflict,
		Code:     "RESOURCE_CONFLICT",
		Message:  fmt.Sprintf("%s already exists: %s", resource, identifier),
		Resource: resource,
	}
}

// DatabaseError creates a database error.
// operation should be a stable identifier (snake_case) for logs and metrics.
func DatabaseError(operation string, err error) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Type:      ErrorTypeDatabase,
		Code:      "DATABASE_ERROR",
		Message:   fmt.Sprintf("database operation failed: %s", operation),
		Operation: operation,
		cause:     err,
	}
}

// ExternalServiceError creates an external service error
func ExternalServiceError(service, operation string, err error) *AppError {
	return &AppError{
		Type:      ErrorTypeExternal,
		Code:      "EXTERNAL_SERVICE_ERROR",
		Message:   fmt.Sprintf("External service error: %s - %s", service, operation),
		Operation: operation,
		Resource:  service,
		cause:     err,
	}
}

// InternalError creates an internal server error
func InternalError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Code:    "INTERNAL_ERROR",
		Message: message,
		cause:   err,
	}
}

// WithDetails creates a copy of the error with details (does not mutate the receiver).
func (e *AppError) WithDetails(details string) *AppError {
	if e == nil {
		return nil
	}
	c := *e
	c.Details = details
	return &c
}

// WithUserMessage creates a copy with a user-facing message.
func (e *AppError) WithUserMessage(message string) *AppError {
	if e == nil {
		return nil
	}
	c := *e
	c.UserMessage = message
	return &c
}

// WithOperation creates a copy with the specified operation.
func (e *AppError) WithOperation(operation string) *AppError {
	if e == nil {
		return nil
	}
	c := *e
	c.Operation = operation
	return &c
}

// WithResource creates a copy with the specified resource.
func (e *AppError) WithResource(resource string) *AppError {
	if e == nil {
		return nil
	}
	c := *e
	c.Resource = resource
	return &c
}

func appErrorAs(err error) (app *AppError, ok bool) {
	if err == nil {
		return nil, false
	}
	if pkgerrs.As(err, &app) {
		return app, true
	}
	return nil, false
}

// IsValidationError checks if the error chain contains a *AppError of validation.
func IsValidationError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeValidation
}

// IsNotFoundError checks if the error chain contains a *AppError not_found.
func IsNotFoundError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeNotFound
}

// IsDatabaseError checks if the error chain contains a *AppError of database.
func IsDatabaseError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeDatabase
}

// GetAppError extracts an AppError from the error
func GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if pkgerrs.As(err, &appErr) {
		return appErr, true
	}

	return nil, false
}

// ServiceUnavailableError creates a service unavailable error
func ServiceUnavailableError(service, reason string) *AppError {
	return &AppError{
		Type:        ErrorTypeServiceUnavailable,
		Code:        "SERVICE_UNAVAILABLE",
		Message:     fmt.Sprintf("%s is temporarily unavailable: %s", service, reason),
		Resource:    service,
		UserMessage: "The service is temporarily unavailable. Please try again later.",
	}
}

// IsServiceUnavailableError checks if the error is a service unavailable error
func IsServiceUnavailableError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeServiceUnavailable
}

// IsUnauthorizedError checks if the error chain contains a *AppError unauthorized.
func IsUnauthorizedError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeUnauthorized
}

// IsForbiddenError checks if the error chain contains a *AppError forbidden.
func IsForbiddenError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeForbidden
}

// IsConflictError checks if the error chain contains a *AppError conflict.
func IsConflictError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeConflict
}

// IsInternalError checks if the error chain contains a *AppError internal.
func IsInternalError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeInternal
}

// IsExternalServiceError checks if the error chain contains a *AppError external.
func IsExternalServiceError(err error) bool {
	app, ok := appErrorAs(err)
	return ok && app.Type == ErrorTypeExternal
}
