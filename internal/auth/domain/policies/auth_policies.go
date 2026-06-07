package policies

import (
	"errors"
	"time"
)

// DefaultSymbols are the default allowed symbols
const DefaultSymbols = "!@#$%^&*()_+-=[]{}|;:,.<>?"

// SessionPolicy defines the session expiration policy
type SessionPolicy struct {
	ExpirationDays int
}

// FailedLoginLockoutPolicy configures automatic account lockout after repeated
// failed local password checks. MaxAttempts 0 disables lockout (counter still increments).
type FailedLoginLockoutPolicy struct {
	MaxAttempts  int
	LockDuration time.Duration
}

const (
	defaultLockoutMaxAttempts = 3
	defaultLockoutMinutes     = 15
)

// DefaultFailedLoginLockout returns the same defaults as AUTH_LOCKOUT_* in .env.example.
func DefaultFailedLoginLockout() FailedLoginLockoutPolicy {
	return FailedLoginLockoutPolicy{
		MaxAttempts:  defaultLockoutMaxAttempts,
		LockDuration: defaultLockoutMinutes * time.Minute,
	}
}

// PasswordPolicy defines the password validation policy
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSymbol    bool
	// AllowedSymbols defines the valid symbols (empty uses DefaultSymbols)
	AllowedSymbols string
}

// Validate validates a password according to the policy
//
//nolint:gocognit
func (p PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return errors.New("password too short")
	}

	if p.RequireUppercase {
		hasUpper := false
		for _, char := range password {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return errors.New("password must contain uppercase letters")
		}
	}

	if p.RequireLowercase {
		hasLower := false
		for _, char := range password {
			if char >= 'a' && char <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return errors.New("password must contain lowercase letters")
		}
	}

	if p.RequireNumber {
		hasNumber := false
		for _, char := range password {
			if char >= '0' && char <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return errors.New("password must contain numbers")
		}
	}

	if p.RequireSymbol {
		symbols := p.AllowedSymbols
		if symbols == "" {
			symbols = DefaultSymbols
		}
		hasSymbol := false
		for _, char := range password {
			for _, symbol := range symbols {
				if char == symbol {
					hasSymbol = true
					break
				}
			}
			if hasSymbol {
				break
			}
		}
		if !hasSymbol {
			return errors.New("password must contain symbols")
		}
	}

	return nil
}
