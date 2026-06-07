package errors

import "errors"

// Domain errors for the users module.
// Use errors.Is() to check the error type.
var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists     = errors.New("email is already in use")
	ErrUsernameAlreadyExists  = errors.New("username is already in use")
	ErrUserIDRequired         = errors.New("user ID is required")
	ErrAtLeastOneUserRequired = errors.New("at least one user must be provided for creation")
	ErrBatchSizeExceeded      = errors.New("user batch exceeds the allowed limit")
	ErrSearchQueryEmpty       = errors.New("search query cannot be empty")
	ErrEmailRequired          = errors.New("email is required")
	ErrPasswordRequired       = errors.New("password is required")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrFirstNameRequired      = errors.New("first name is required")
	ErrLastNameRequired       = errors.New("last name is required")
	ErrRoleNameRequired       = errors.New("role name is required")
)
