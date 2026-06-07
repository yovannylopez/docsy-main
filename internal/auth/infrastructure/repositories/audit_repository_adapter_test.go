package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	infraTypes "github.com/yovannylopez/docsy-main/internal/auth/infrastructure/types"
)

func TestNewAuditRepositoryAdapter(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	require.NotNil(t, r)
}

func TestAuditRepositoryAdapter_LogAction_NilLog(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))

	require.Error(t, r.LogAction(context.Background(), nil))
}

func TestAuditRepositoryAdapter_LogAction_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	log := repoAuditLogForInsert("log-1", domain.AuditActionCreate, domain.AuditResultSuccess, time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC), []string{"f1"})

	mock.ExpectExec(regexp.MustCompile(`INSERT INTO audit_logs`).String()).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, r.LogAction(context.Background(), log))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_LogAction_WithJSONMaps(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	prev := map[string]any{"a": 1}
	nw := map[string]any{"b": 2}
	log := repoAuditLogWithJSONMaps("log-2", time.Now().UTC(), prev, nw)

	mock.ExpectExec(regexp.MustCompile(`INSERT INTO audit_logs`).String()).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, r.LogAction(context.Background(), log))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_LogAction_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	log := repoAuditLogForInsert("x", "a", "ok", time.Now().UTC(), nil)

	mock.ExpectExec(regexp.MustCompile(`INSERT INTO audit_logs`).String()).
		WillReturnError(errors.New("insert failed"))

	require.Error(t, r.LogAction(context.Background(), log))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_GetUserAuditLogs_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	cols := []string{
		"id", "user_id", "session_id", "action", "resource", "resource_id",
		"result", "message", "ip_address", "user_agent", "request_id",
		"previous_data", "new_data", "changed_fields", "created_at",
	}
	mock.ExpectQuery(regexp.MustCompile(`FROM audit_logs`).String()).
		WithArgs("u1", 10, 0).
		WillReturnRows(sqlmock.NewRows(cols))

	out, err := r.GetUserAuditLogs(context.Background(), "u1", 10, 0)
	require.NoError(t, err)
	assert.Empty(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_GetUserAuditLogs_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`FROM audit_logs`).String()).
		WillReturnError(errors.New("db"))

	_, err = r.GetUserAuditLogs(context.Background(), "u", 1, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_GetSessionAuditLogs_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	cols := []string{
		"id", "user_id", "session_id", "action", "resource", "resource_id",
		"result", "message", "ip_address", "user_agent", "request_id",
		"previous_data", "new_data", "changed_fields", "created_at",
	}
	mock.ExpectQuery(regexp.MustCompile(`WHERE session_id`).String()).
		WithArgs("s1", 5, 0).
		WillReturnRows(sqlmock.NewRows(cols))

	out, err := r.GetSessionAuditLogs(context.Background(), "s1", 5, 0)
	require.NoError(t, err)
	assert.Empty(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_List_NoFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	f := &dtos.AuditLogFilters{Limit: 20, Offset: 0}

	mock.ExpectQuery(regexp.MustCompile(`SELECT COUNT\(\*\) FROM audit_logs`).String()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	cols := []string{
		"id", "user_id", "session_id", "action", "resource", "resource_id",
		"result", "message", "ip_address", "user_agent", "request_id",
		"previous_data", "new_data", "changed_fields", "created_at",
	}
	now := time.Now().UTC()
	mock.ExpectQuery(regexp.MustCompile(`ORDER BY created_at DESC`).String()).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(
			"l1", nil, nil, domain.AuditActionRead, nil, nil, domain.AuditResultSuccess, nil, nil, nil, nil,
			nil, nil, []byte("{}"), now,
		))

	logs, total, err := r.List(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, logs, 1)
	assert.Equal(t, "l1", logs[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	f := &dtos.AuditLogFilters{Limit: 10, Offset: 0}

	mock.ExpectQuery(regexp.MustCompile(`SELECT COUNT\(\*\) FROM audit_logs`).String()).
		WillReturnError(errors.New("count fail"))

	_, _, err = r.List(context.Background(), f)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_List_WithUserFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	r := NewAuditRepositoryAdapter(sqlx.NewDb(db, "postgres"))
	uid := "user-9"
	f := &dtos.AuditLogFilters{UserID: &uid, Limit: 5, Offset: 1}

	mock.ExpectQuery(regexp.MustCompile(`SELECT COUNT\(\*\) FROM audit_logs`).String()).
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	cols := []string{
		"id", "user_id", "session_id", "action", "resource", "resource_id",
		"result", "message", "ip_address", "user_agent", "request_id",
		"previous_data", "new_data", "changed_fields", "created_at",
	}
	mock.ExpectQuery(regexp.MustCompile(`ORDER BY created_at DESC`).String()).
		WithArgs(uid, 5, 1).
		WillReturnRows(sqlmock.NewRows(cols))

	logs, total, err := r.List(context.Background(), f)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, logs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditRepositoryAdapter_convertRowsToAuditLogs_JSON(t *testing.T) {
	r := &AuditRepositoryAdapter{}
	now := time.Now().UTC()
	j := infraTypes.JSONB{"x": true}
	row := AuditLogRow{
		ID:            "id",
		Action:        "act",
		Result:        domain.AuditResultSuccess,
		CreatedAt:     now,
		PreviousData:  &j,
		NewData:       &j,
		ChangedFields: nil,
	}
	out := r.convertRowsToAuditLogs([]AuditLogRow{row})
	require.Len(t, out, 1)
	require.NotNil(t, out[0].PreviousData)
	require.NotNil(t, out[0].NewData)
}
