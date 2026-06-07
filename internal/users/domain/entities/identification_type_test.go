package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateIdentificationType_NilOrEmpty(t *testing.T) {
	assert.NoError(t, ValidateIdentificationType(nil))

	empty := ""
	assert.NoError(t, ValidateIdentificationType(&empty))
}

func TestValidateIdentificationType_ValidLowercase(t *testing.T) {
	for _, code := range []string{"cc", "ce", "pa", "nit", "rut"} {
		s := code
		require.NoError(t, ValidateIdentificationType(&s), code)
	}
}

func TestValidateIdentificationType_ValidMixedCase(t *testing.T) {
	s := "  CC  "
	require.NoError(t, ValidateIdentificationType(&s))
}

func TestValidateIdentificationType_Invalid(t *testing.T) {
	s := "dni"
	err := ValidateIdentificationType(&s)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidIdentificationType)
}
