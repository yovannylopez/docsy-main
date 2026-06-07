package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	h := NewPasswordHasher()
	require.NotNil(t, h)

	hash, err := h.HashPassword("correct-horse-battery-staple")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "correct-horse-battery-staple", hash)

	ok, err := h.VerifyPassword("correct-horse-battery-staple", hash)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = h.VerifyPassword("wrong-password", hash)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestPasswordHasher_VerifyPassword_InvalidHash(t *testing.T) {
	h := NewPasswordHasher()
	ok, err := h.VerifyPassword("password", "not-a-bcrypt-hash")
	require.NoError(t, err)
	assert.False(t, ok)
}
