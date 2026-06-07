package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenFingerprint_Deterministic(t *testing.T) {
	tok := "refresh-abc"
	a := RefreshTokenFingerprint(tok)
	b := RefreshTokenFingerprint(tok)
	assert.Equal(t, a, b)
	assert.Len(t, a, 64)
}
