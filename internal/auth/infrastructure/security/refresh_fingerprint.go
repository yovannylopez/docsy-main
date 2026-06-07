package security

import (
	"crypto/sha256"
	"encoding/hex"
)

// RefreshTokenFingerprint returns the SHA-256 hex digest of the refresh token (for lookup/validation in sessions).
func RefreshTokenFingerprint(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}
