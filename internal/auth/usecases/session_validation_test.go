package usecases

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

func TestValidateActiveSession(t *testing.T) {
	t.Parallel()

	now := time.Now()
	valid := &entities.Session{
		UserID:    "u1",
		IsActive:  true,
		ExpiresAt: now.Add(time.Hour),
	}

	t.Run("valid session", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validateActiveSession(valid, "u1"))
	})

	t.Run("nil session", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, validateActiveSession(nil, "u1"))
	})

	t.Run("user mismatch", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, validateActiveSession(valid, "other"))
	})

	t.Run("skip user check when userID empty", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validateActiveSession(valid, ""))
	})

	t.Run("revoked", func(t *testing.T) {
		t.Parallel()
		revoked := *valid
		revoked.IsActive = false
		assert.Error(t, validateActiveSession(&revoked, "u1"))
	})

	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		expired := *valid
		expired.ExpiresAt = now.Add(-time.Hour)
		assert.Error(t, validateActiveSession(&expired, "u1"))
	})
}
