package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

func TestNewNoOpAuditRepository(t *testing.T) {
	r := NewNoOpAuditRepository()
	require.NotNil(t, r)
}

func TestNoOpAuditRepository_AllMethods(t *testing.T) {
	r := NewNoOpAuditRepository()
	ctx := context.Background()

	require.NoError(t, r.LogAction(ctx, repoAuditLogForInsert("", "x", "ok", time.Now(), nil)))

	logs, err := r.GetUserAuditLogs(ctx, "user-1", 10, 0)
	require.NoError(t, err)
	require.Empty(t, logs)

	slogs, err := r.GetSessionAuditLogs(ctx, "sess-1", 5, 0)
	require.NoError(t, err)
	require.Empty(t, slogs)

	list, total, err := r.List(ctx, &dtos.AuditLogFilters{UserID: authtest.StringPtr("u"), Limit: 20, Offset: 0})
	require.NoError(t, err)
	require.Empty(t, list)
	require.Zero(t, total)
}
