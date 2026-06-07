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

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

func TestNewPasswordHistoryRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))
	require.NotNil(t, repo)
}

func TestPasswordHistoryRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	ph := &entities.PasswordHistory{
		UserID:       "uid-1",
		PasswordHash: "hash-old",
		ChangedAt:    now,
		ChangedBy:    nil,
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO password_history`)).
		WithArgs(ph.UserID, ph.PasswordHash, ph.ChangedAt, ph.ChangedBy).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.Create(context.Background(), ph))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordHistoryRepository_Create_SetsChangedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))
	ph := &entities.PasswordHistory{
		UserID:       "uid-1",
		PasswordHash: "hash-old",
		// ChangedAt intentionally zero → should be set by Create
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO password_history`)).
		WithArgs(ph.UserID, ph.PasswordHash, sqlmock.AnyArg(), ph.ChangedBy).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.Create(context.Background(), ph))
	assert.False(t, ph.ChangedAt.IsZero(), "ChangedAt should be set when zero")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordHistoryRepository_Create_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))
	ph := &entities.PasswordHistory{
		UserID:       "uid-1",
		PasswordHash: "hash-old",
		ChangedAt:    time.Now(),
	}

	dbErr := errors.New("connection reset")
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO password_history`)).
		WithArgs(ph.UserID, ph.PasswordHash, ph.ChangedAt, ph.ChangedBy).
		WillReturnError(dbErr)

	err = repo.Create(context.Background(), ph)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password_history create")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordHistoryRepository_GetUserPasswordHistory_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))
	userID := "uid-1"
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "user_id", "password_hash", "changed_at", "changed_by"}).
		AddRow("entry-1", userID, "hash-a", now, nil).
		AddRow("entry-2", userID, "hash-b", now.Add(-time.Hour), nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, password_hash, changed_at, changed_by`)).
		WithArgs(userID, 5).
		WillReturnRows(rows)

	result, err := repo.GetUserPasswordHistory(context.Background(), userID, 5)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "entry-1", result[0].ID)
	assert.Equal(t, "hash-a", result[0].PasswordHash)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordHistoryRepository_GetUserPasswordHistory_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))

	rows := sqlmock.NewRows([]string{"id", "user_id", "password_hash", "changed_at", "changed_by"})
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, password_hash, changed_at, changed_by`)).
		WithArgs("uid-x", 5).
		WillReturnRows(rows)

	result, err := repo.GetUserPasswordHistory(context.Background(), "uid-x", 5)
	require.NoError(t, err)
	assert.Empty(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordHistoryRepository_GetUserPasswordHistory_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, password_hash, changed_at, changed_by`)).
		WithArgs("uid-1", 5).
		WillReturnError(errors.New("timeout"))

	_, err = repo.GetUserPasswordHistory(context.Background(), "uid-1", 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password_history get history")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordHistoryRepository_CheckPasswordInHistory_AlwaysFalse(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))
	found, err := repo.CheckPasswordInHistory(context.Background(), "uid-1", "anyhash")
	require.NoError(t, err)
	assert.False(t, found, "CheckPasswordInHistory must always return false (bcrypt not comparable at SQL level)")
}

func TestPasswordHistoryRepository_CleanOldPasswordHistory_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM password_history`)).
		WithArgs("uid-1", 5).
		WillReturnResult(sqlmock.NewResult(0, 2))

	require.NoError(t, repo.CleanOldPasswordHistory(context.Background(), "uid-1", 5))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordHistoryRepository_CleanOldPasswordHistory_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPasswordHistoryRepository(sqlx.NewDb(db, "postgres"))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM password_history`)).
		WithArgs("uid-1", 5).
		WillReturnError(errors.New("db down"))

	err = repo.CleanOldPasswordHistory(context.Background(), "uid-1", 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password_history clean")
	require.NoError(t, mock.ExpectationsWereMet())
}
