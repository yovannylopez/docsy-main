package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	apperrors "github.com/yovannylopez/docsy-main/pkg/errors"
)

// VerificationTokenRepository implements ports.VerificationTokenRepository using sqlx.
type VerificationTokenRepository struct {
	db *sqlx.DB
}

// NewVerificationTokenRepository creates a new VerificationTokenRepository.
func NewVerificationTokenRepository(db *sqlx.DB) ports.VerificationTokenRepository {
	return &VerificationTokenRepository{db: db}
}

// CreateToken inserts a new verification token record.
func (r *VerificationTokenRepository) CreateToken(ctx context.Context, token *entities.VerificationToken) error {
	query := `
		INSERT INTO verification_tokens (id, user_id, token_hash, token_type, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.TokenType,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return apperrors.DatabaseError("create_verification_token", err)
	}
	return nil
}

// FindTokenByHash looks up a token by its hash value.
// Returns sql.ErrNoRows (wrapped) when not found.
func (r *VerificationTokenRepository) FindTokenByHash(
	ctx context.Context,
	tokenHash string,
) (*entities.VerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_type, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE token_hash = $1
	`
	var t entities.VerificationToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.TokenType,
		&t.ExpiresAt,
		&t.UsedAt,
		&t.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("verification_token", tokenHash)
		}
		return nil, apperrors.DatabaseError("find_verification_token_by_hash", err)
	}
	return &t, nil
}

// MarkTokenAsUsed stamps the used_at timestamp on a token record.
func (r *VerificationTokenRepository) MarkTokenAsUsed(ctx context.Context, tokenID string) error {
	query := `
		UPDATE verification_tokens
		SET used_at = $2
		WHERE id = $1
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, tokenID, now)
	if err != nil {
		return apperrors.DatabaseError("mark_verification_token_used", err)
	}
	return nil
}

// CleanExpiredTokens removes all tokens that have expired and have been used.
func (r *VerificationTokenRepository) CleanExpiredTokens(ctx context.Context) error {
	query := `
		DELETE FROM verification_tokens
		WHERE expires_at < NOW() AND used_at IS NOT NULL
	`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return apperrors.DatabaseError("clean_expired_verification_tokens", err)
	}
	return nil
}

// GetUserTokens returns all tokens of a given type for a user, ordered by creation date descending.
func (r *VerificationTokenRepository) GetUserTokens(
	ctx context.Context,
	userID, tokenType string,
) ([]entities.VerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_type, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE user_id = $1 AND token_type = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, tokenType)
	if err != nil {
		return nil, apperrors.DatabaseError("get_user_verification_tokens", err)
	}
	defer func() { _ = rows.Close() }()

	var tokens []entities.VerificationToken
	for rows.Next() {
		var t entities.VerificationToken
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.TokenHash,
			&t.TokenType,
			&t.ExpiresAt,
			&t.UsedAt,
			&t.CreatedAt,
		); err != nil {
			return nil, apperrors.DatabaseError("scan_verification_token", err)
		}
		tokens = append(tokens, t)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("iterate_verification_tokens", err)
	}
	return tokens, nil
}
