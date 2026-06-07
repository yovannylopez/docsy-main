package domain

import "errors"

// Sentinel errors for the change-password flow.
// Defined at domain level so that both use cases (which produce them) and
// transport (which maps them to HTTP responses) can import this package
// without creating a usecases → transport or transport → usecases dependency.
var (
	// ErrCurrentPasswordInvalid is returned when current password verification fails.
	// Intentionally generic to avoid leaking credential information (FR-002).
	ErrCurrentPasswordInvalid = errors.New("current password invalid")

	// ErrSamePassword is returned when the new password equals the current one.
	ErrSamePassword = errors.New("new password must differ from current password")

	// ErrPasswordInHistory is returned when the new password matches one of the
	// user's recent passwords stored in password_history (PH-FR-002).
	ErrPasswordInHistory = errors.New("new password was recently used")

	// MFA sentinel errors.

	// ErrMFAAlreadyEnabled is returned when a user attempts to set up MFA but it is already active.
	ErrMFAAlreadyEnabled = errors.New("MFA already enabled")

	// ErrMFANotEnabled is returned when an operation requires MFA to be active but it is not.
	ErrMFANotEnabled = errors.New("MFA not enabled")

	// ErrMFAInvalidCode is returned when the supplied TOTP code is invalid or expired.
	ErrMFAInvalidCode = errors.New("invalid or expired MFA code")

	// ErrMFAInvalidToken is returned when a setup or challenge token is invalid, expired, or already used.
	ErrMFAInvalidToken = errors.New("invalid or expired MFA token")

	// ErrEmailAlreadyExists is returned when signup targets an email that is already registered.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrUsernameAlreadyExists is returned when the chosen username is already taken.
	ErrUsernameAlreadyExists = errors.New("username already exists")

	// ErrUsernameTooLong is returned when the trimmed username exceeds the allowed length.
	ErrUsernameTooLong = errors.New("username too long")
)
