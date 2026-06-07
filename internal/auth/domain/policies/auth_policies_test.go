package policies

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func defaultTestPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSymbol:    true,
		AllowedSymbols:   DefaultSymbols,
	}
}

func TestPasswordPolicy_Validate_TooShort(t *testing.T) {
	policy := defaultTestPolicy()
	err := policy.Validate("Ab1!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password too short")
}

func TestPasswordPolicy_Validate_NoUppercase(t *testing.T) {
	policy := defaultTestPolicy()
	err := policy.Validate("weak1!pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uppercase")
}

func TestPasswordPolicy_Validate_NoLowercase(t *testing.T) {
	policy := defaultTestPolicy()
	err := policy.Validate("WEAK1!PASS")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase")
}

func TestPasswordPolicy_Validate_NoNumber(t *testing.T) {
	policy := defaultTestPolicy()
	err := policy.Validate("Weak!Pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "numbers")
}

func TestPasswordPolicy_Validate_NoSymbol(t *testing.T) {
	policy := defaultTestPolicy()
	err := policy.Validate("Weak1Pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symbols")
}

func TestPasswordPolicy_Validate_ValidPassword(t *testing.T) {
	policy := defaultTestPolicy()
	err := policy.Validate("Str0ng!Pass")
	assert.NoError(t, err)
}

func TestPasswordPolicy_Validate_RequireUppercaseFalse(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:        8,
		RequireUppercase: false,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSymbol:    true,
		AllowedSymbols:   DefaultSymbols,
	}
	err := policy.Validate("weak1!pass")
	assert.NoError(t, err)
}

func TestPasswordPolicy_Validate_EmptyAllowedSymbolsUsesDefault(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSymbol:    true,
		AllowedSymbols:   "",
	}
	err := policy.Validate("Str0ng!Pass")
	assert.NoError(t, err)
}

func TestDefaultFailedLoginLockout(t *testing.T) {
	p := DefaultFailedLoginLockout()
	assert.Equal(t, 3, p.MaxAttempts)
	assert.Equal(t, 15*time.Minute, p.LockDuration)
}
