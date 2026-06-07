package repositories

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

func TestNewVerificationTokenRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	require.NotNil(t, repo)
}

func TestVerificationTokenRepository_CreateToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	token := &entities.VerificationToken{
		ID:        uuid.NewString(),
		UserID:    uuid.NewString(),
		TokenHash: "sha256hashvalue",
		TokenType: domain.VerificationTokenTypeMFASetup,
		ExpiresAt: now.Add(10 * time.Minute),
		CreatedAt: now,
	}

	mock.ExpectExec(regexp.MustCompile(`INSERT INTO verification_tokens`).String()).
		WithArgs(token.ID, token.UserID, token.TokenHash, token.TokenType, token.ExpiresAt, token.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.CreateToken(context.Background(), token))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_CreateToken_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	token := &entities.VerificationToken{
		ID:        uuid.NewString(),
		UserID:    uuid.NewString(),
		TokenHash: "hashval",
		TokenType: domain.VerificationTokenTypeMFASetup,
		ExpiresAt: now.Add(time.Minute),
		CreatedAt: now,
	}

	mock.ExpectExec(regexp.MustCompile(`INSERT INTO verification_tokens`).String()).
		WillReturnError(errors.New("db error"))

	require.Error(t, repo.CreateToken(context.Background(), token))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_FindTokenByHash_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	tokenID := uuid.NewString()
	userID := uuid.NewString()
	hash := "sha256hashvalue_abc"
	expires := now.Add(time.Hour)

	cols := []string{"id", "user_id", "token_hash", "token_type", "expires_at", "used_at", "created_at"}
	mock.ExpectQuery(regexp.MustCompile(`SELECT .* FROM verification_tokens WHERE token_hash`).String()).
		WithArgs(hash).
		WillReturnRows(
			sqlmock.NewRows(cols).AddRow(tokenID, userID, hash, domain.VerificationTokenTypeMFAChallenge, expires, nil, now),
		)

	result, err := repo.FindTokenByHash(context.Background(), hash)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, tokenID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, hash, result.TokenHash)
	assert.Equal(t, domain.VerificationTokenTypeMFAChallenge, result.TokenType)
	assert.Nil(t, result.UsedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_FindTokenByHash_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))

	mock.ExpectQuery(regexp.MustCompile(`SELECT .* FROM verification_tokens WHERE token_hash`).String()).
		WithArgs("nonexistent_hash").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.FindTokenByHash(context.Background(), "nonexistent_hash")

	require.Error(t, err)
	assert.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_FindTokenByHash_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))

	mock.ExpectQuery(regexp.MustCompile(`SELECT .* FROM verification_tokens WHERE token_hash`).String()).
		WithArgs("any_hash").
		WillReturnError(errors.New("connection reset"))

	result, err := repo.FindTokenByHash(context.Background(), "any_hash")

	require.Error(t, err)
	assert.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_MarkTokenAsUsed_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	tokenID := uuid.NewString()

	mock.ExpectExec(regexp.MustCompile(`UPDATE verification_tokens`).String()).
		WithArgs(tokenID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.MarkTokenAsUsed(context.Background(), tokenID))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_MarkTokenAsUsed_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	tokenID := uuid.NewString()

	mock.ExpectExec(regexp.MustCompile(`UPDATE verification_tokens`).String()).
		WillReturnError(errors.New("update failed"))

	require.Error(t, repo.MarkTokenAsUsed(context.Background(), tokenID))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_GetUserTokens_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	now := time.Now().UTC()
	userID := uuid.NewString()
	tokenType := domain.VerificationTokenTypeMFASetup

	cols := []string{"id", "user_id", "token_hash", "token_type", "expires_at", "used_at", "created_at"}
	mock.ExpectQuery(regexp.MustCompile(`SELECT .* FROM verification_tokens WHERE user_id`).String()).
		WithArgs(userID, tokenType).
		WillReturnRows(
			sqlmock.NewRows(cols).
				AddRow(uuid.NewString(), userID, "hash1", tokenType, now.Add(time.Hour), nil, now).
				AddRow(uuid.NewString(), userID, "hash2", tokenType, now.Add(time.Hour), nil, now),
		)

	tokens, err := repo.GetUserTokens(context.Background(), userID, tokenType)

	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_GetUserTokens_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	userID := uuid.NewString()
	tokenType := domain.VerificationTokenTypeMFAChallenge

	cols := []string{"id", "user_id", "token_hash", "token_type", "expires_at", "used_at", "created_at"}
	mock.ExpectQuery(regexp.MustCompile(`SELECT .* FROM verification_tokens WHERE user_id`).String()).
		WithArgs(userID, tokenType).
		WillReturnRows(sqlmock.NewRows(cols))

	tokens, err := repo.GetUserTokens(context.Background(), userID, tokenType)

	require.NoError(t, err)
	assert.Empty(t, tokens)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerificationTokenRepository_GetUserTokens_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerificationTokenRepository(sqlx.NewDb(db, "postgres"))
	userID := uuid.NewString()

	mock.ExpectQuery(regexp.MustCompile(`SELECT .* FROM verification_tokens WHERE user_id`).String()).
		WillReturnError(errors.New("query failed"))

	tokens, err := repo.GetUserTokens(context.Background(), userID, domain.VerificationTokenTypeMFASetup)

	require.Error(t, err)
	assert.Nil(t, tokens)
	require.NoError(t, mock.ExpectationsWereMet())
}
