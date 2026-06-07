package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetTotalUsersCount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewUserRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	n, err := repo.GetTotalUsersCount(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 7, n)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetTotalUsersCount_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewUserRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).
		WillReturnError(errors.New("db error"))

	_, err = repo.GetTotalUsersCount(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
