package repositories

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	require.NotNil(t, repo)
}

func TestSessionRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	sess := repoSessionForCreate("sid", "uid", "hash", now)

	mock.ExpectExec(regexp.MustCompile(`INSERT INTO sessions`).String()).
		WithArgs(
			sess.ID, sess.UserID, sess.RefreshTokenHash, sess.AccessTokenJTI, sess.UserAgent,
			sess.IPAddress, sess.Location, sess.DeviceFingerprint, sess.CreatedAt, sess.LastUsedAt,
			sess.ExpiresAt, sess.IsActive,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.Create(context.Background(), sess))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Create_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	sess := repoSessionForCreate("s", "u", "h", now)

	mock.ExpectExec(regexp.MustCompile(`INSERT INTO sessions`).String()).
		WillReturnError(errors.New("exec failed"))

	require.Error(t, repo.Create(context.Background(), sess))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`SELECT id, user_id`).String()).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	out, err := repo.FindByID(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "refresh_token_hash", "access_token_jti", "user_agent",
		"ip_address", "location", "device_fingerprint", "created_at", "last_used_at",
		"expires_at", "is_active", "revoked_at", "revoked_reason",
	}).AddRow("sid", "uid", "rh", nil, nil, nil, nil, nil, now, now, now, true, nil, nil)

	mock.ExpectQuery(regexp.MustCompile(`SELECT id, user_id`).String()).
		WithArgs("sid").
		WillReturnRows(rows)

	out, err := repo.FindByID(context.Background(), "sid")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "sid", out.ID)
	assert.Equal(t, "uid", out.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`SELECT id, user_id`).String()).
		WithArgs("x").
		WillReturnError(errors.New("db boom"))

	_, err = repo.FindByID(context.Background(), "x")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	r := sqlmock.NewRows([]string{
		"id", "user_id", "refresh_token_hash", "access_token_jti", "user_agent",
		"ip_address", "location", "device_fingerprint", "created_at", "last_used_at",
		"expires_at", "is_active", "revoked_at", "revoked_reason",
	}).AddRow("s1", "u1", "h", nil, nil, nil, nil, nil, now, now, now, true, nil, nil)

	mock.ExpectQuery(regexp.MustCompile(`FROM sessions`).String()).
		WithArgs("u1").
		WillReturnRows(r)

	list, err := repo.FindByUserID(context.Background(), "u1")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "s1", list[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByRefreshToken_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`refresh_token_hash`).String()).
		WithArgs("hashx").
		WillReturnError(sql.ErrNoRows)

	out, err := repo.FindByRefreshToken(context.Background(), "hashx")
	require.NoError(t, err)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	sess := repoSessionForCreate("sid", "u", "r", now)

	mock.ExpectExec(regexp.MustCompile(`UPDATE sessions`).String()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Update(context.Background(), sess))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_UpdateLastUsed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectExec(regexp.MustCompile(`UPDATE sessions`).String()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateLastUsed(context.Background(), "sid"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_RevokeSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectExec(regexp.MustCompile(`UPDATE sessions`).String()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.RevokeSession(context.Background(), "sid", "reason"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_RevokeAllUserSessions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectExec(regexp.MustCompile(`UPDATE sessions`).String()).
		WillReturnResult(sqlmock.NewResult(0, 2))
	require.NoError(t, repo.RevokeAllUserSessions(context.Background(), "u", "all"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_CleanupExpiredSessions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectExec(regexp.MustCompile(`UPDATE sessions`).String()).
		WillReturnResult(sqlmock.NewResult(0, 3))
	require.NoError(t, repo.CleanupExpiredSessions(context.Background()))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByUserID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSessionRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`FROM sessions`).String()).
		WillReturnError(errors.New("q"))

	_, err = repo.FindByUserID(context.Background(), "u")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
