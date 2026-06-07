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

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

func userSelectColumns() []string {
	return []string{
		"id", "email", "username", "password_hash", "first_name", "last_name",
		"identification_number", "identification_type", "phone",
		"is_active", "is_verified", "last_login_at", "failed_login_attempts",
		"last_failed_login_at", "locked_until", "mfa_enabled", "mfa_secret",
		"password_changed_at", "must_change_password", "created_at", "updated_at",
		"created_by", "updated_by",
	}
}

func TestUserRepository_Create_Success_NoRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	u := repoUserForCreate(now)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.MustCompile(`INSERT INTO users`).String()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.Create(context.Background(), u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	u := repoUserForCreate(now)

	mock.ExpectBegin().WillReturnError(errors.New("begin"))
	require.Error(t, repo.Create(context.Background(), u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_InsertError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	u := repoUserForCreate(now)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.MustCompile(`INSERT INTO users`).String()).WillReturnError(errors.New("ins"))
	mock.ExpectRollback()

	require.Error(t, repo.Create(context.Background(), u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_WithRoles(t *testing.T) {
	require.NoError(t, logging.Init(false))

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	rid := "role-1"
	u := repoUserForCreate(now)
	u.Roles = []entities.Role{{ID: rid}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.MustCompile(`INSERT INTO users`).String()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.MustCompile(`INSERT INTO user_roles`).String()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.Create(context.Background(), u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`FROM users u`).String()).
		WithArgs("missing@test.co").
		WillReturnError(sql.ErrNoRows)

	u, err := repo.FindByEmail(context.Background(), "missing@test.co")
	require.NoError(t, err)
	assert.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_Found_WithRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	uid := "user-1"
	rows := sqlmock.NewRows(userSelectColumns()).AddRow(
		uid, "e@test.co", nil, "h", "F", "L", nil, nil, nil,
		true, false, nil, 0, nil, nil, false, nil, now, false, now, now, nil, nil,
	)
	mock.ExpectQuery(regexp.MustCompile(`FROM users u`).String()).
		WithArgs("e@test.co").
		WillReturnRows(rows)

	roleCols := []string{"id", "name", "description", "is_system_role", "is_active", "created_at", "updated_at"}
	roleRows := sqlmock.NewRows(roleCols).AddRow("r1", "admin", nil, true, true, now, now)
	mock.ExpectQuery(regexp.MustCompile(`FROM roles r`).String()).
		WithArgs(uid).
		WillReturnRows(roleRows)

	mock.ExpectQuery(regexp.MustCompile(`FROM permissions p`).String()).
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"name"}))

	u, err := repo.FindByEmail(context.Background(), "e@test.co")
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, uid, u.ID)
	require.Len(t, u.Roles, 1)
	assert.Equal(t, "admin", u.Roles[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`FROM users u`).String()).
		WithArgs("id-x").
		WillReturnError(sql.ErrNoRows)

	u, err := repo.FindByID(context.Background(), "id-x")
	require.NoError(t, err)
	assert.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`u\.username\s*=\s*\$1`).String()).
		WithArgs("jdoe").
		WillReturnError(sql.ErrNoRows)

	u, err := repo.FindByUsername(context.Background(), "jdoe")
	require.NoError(t, err)
	assert.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	u := repoUserForUpdateSuccess(now)
	mock.ExpectExec(regexp.MustCompile(`UPDATE users`).String()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Update(context.Background(), u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	u := repoUserForUpdateNoRows(now)
	mock.ExpectExec(regexp.MustCompile(`UPDATE users`).String()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	require.Error(t, repo.Update(context.Background(), u))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetRoleByName_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`FROM roles`).String()).
		WithArgs("nope").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetRoleByName(context.Background(), "nope")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetRoleByName_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	r := sqlmock.NewRows([]string{"id", "name", "description", "is_system_role", "is_active", "created_at", "updated_at"}).
		AddRow("r1", "admin", nil, true, true, now, now)
	mock.ExpectQuery(regexp.MustCompile(`FROM roles`).String()).
		WithArgs("admin").
		WillReturnRows(r)

	role, err := repo.GetRoleByName(context.Background(), "admin")
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, "admin", role.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SimpleExecMethods(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	until := time.Now().UTC()

	mock.ExpectExec(regexp.MustCompile(`UPDATE users`).String()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.MustCompile(`RETURNING failed_login_attempts`).String()).
		WithArgs("u1", 0, int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"failed_login_attempts", "locked_until"}).AddRow(1, nil))
	for i := 0; i < 8; i++ {
		mock.ExpectExec(regexp.MustCompile(`UPDATE users`).String()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	require.NoError(t, repo.UpdateLastLogin(context.Background(), "u1"))
	require.NoError(t, repo.IncrementFailedLoginAttempts(context.Background(), "u1"))
	require.NoError(t, repo.ResetFailedLoginAttempts(context.Background(), "u1"))
	require.NoError(t, repo.LockUserAccount(context.Background(), "u1", &until))
	require.NoError(t, repo.UnlockUserAccount(context.Background(), "u1"))
	require.NoError(t, repo.UpdatePassword(context.Background(), "u1", "hash"))
	require.NoError(t, repo.SetMustChangePassword(context.Background(), "u1", true))
	require.NoError(t, repo.EnableMFA(context.Background(), "u1", "sec"))
	require.NoError(t, repo.DisableMFA(context.Background(), "u1"))
	require.NoError(t, repo.VerifyUser(context.Background(), "u1"))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetAllUsers_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	r := sqlmock.NewRows(userSelectColumns()).AddRow(
		"u1", "a@b.co", nil, "h", "F", "L", nil, nil, nil,
		true, false, nil, 0, nil, nil, false, nil, now, false, now, now, nil, nil,
	)
	mock.ExpectQuery(regexp.MustCompile(`FROM users u`).String()).
		WithArgs(10, 0).
		WillReturnRows(r)

	list, err := repo.GetAllUsers(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "u1", list[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetAllUsers_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`FROM users u`).String()).
		WillReturnError(errors.New("q"))

	_, err = repo.GetAllUsers(context.Background(), 5, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SearchUsers_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	r := sqlmock.NewRows(userSelectColumns()).AddRow(
		"u1", "find@me.co", nil, "h", "F", "L", nil, nil, nil,
		true, false, nil, 0, nil, nil, false, nil, now, false, now, now, nil, nil,
	)
	mock.ExpectQuery(regexp.MustCompile(`to_tsvector\('spanish', u.email\)`).String()).
		WithArgs("find", "%find%", 5, 0).
		WillReturnRows(r)

	list, err := repo.SearchUsers(context.Background(), "find", nil, 5, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SearchUsers_WithActivo(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	act := true
	r := sqlmock.NewRows(userSelectColumns()).AddRow(
		"u1", "find@me.co", nil, "h", "F", "L", nil, nil, nil,
		true, false, nil, 0, nil, nil, false, nil, now, false, now, now, nil, nil,
	)
	mock.ExpectQuery(regexp.MustCompile(`to_tsvector\('spanish', u.email\)`).String()).
		WithArgs("find", "%find%", true, 5, 0).
		WillReturnRows(r)

	list, err := repo.SearchUsers(context.Background(), "find", &act, 5, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CountSearchUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`SELECT COUNT\(\*\)`).String()).
		WithArgs("q", "%q%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

	n, err := repo.CountSearchUsers(context.Background(), "q", nil)
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CountSearchUsers_WithActivo(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	act := false
	mock.ExpectQuery(regexp.MustCompile(`SELECT COUNT\(\*\)`).String()).
		WithArgs("q", "%q%", false).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	n, err := repo.CountSearchUsers(context.Background(), "q", &act)
	require.NoError(t, err)
	assert.Zero(t, n)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_RecordFailedPasswordAttempt_NoLockout(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	mock.ExpectQuery(regexp.MustCompile(`RETURNING failed_login_attempts`).String()).
		WithArgs("u1", 0, int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"failed_login_attempts", "locked_until"}).AddRow(2, nil))

	res, err := repo.RecordFailedPasswordAttempt(context.Background(), "u1", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, res.FailedAttempts)
	assert.Nil(t, res.LockedUntil)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_RecordFailedPasswordAttempt_SetsLockAtThreshold(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	lu := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.MustCompile(`RETURNING failed_login_attempts`).String()).
		WithArgs("u1", 3, int64(900)).
		WillReturnRows(sqlmock.NewRows([]string{"failed_login_attempts", "locked_until"}).AddRow(3, lu))

	res, err := repo.RecordFailedPasswordAttempt(context.Background(), "u1", 3, 15*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 3, res.FailedAttempts)
	require.NotNil(t, res.LockedUntil)
	assert.True(t, res.LockedUntil.Equal(lu))
	require.NoError(t, mock.ExpectationsWereMet())
}
