package domain

import "errors"

// ErrInvalidCredentials is returned when email/password verification fails.
var ErrInvalidCredentials = errors.New("invalid credentials")
