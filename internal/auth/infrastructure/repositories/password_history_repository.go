package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// PasswordHistoryRepository is the sqlx-based implementation for password history operations.
type PasswordHistoryRepository struct {
	db *sqlx.DB
}

// NewPasswordHistoryRepository creates a new instance of PasswordHistoryRepository.
func NewPasswordHistoryRepository(db *sqlx.DB) *PasswordHistoryRepository {
	return &PasswordHistoryRepository{db: db}
}

// Create inserts a new password history record for a user.
func (r *PasswordHistoryRepository) Create(ctx context.Context, ph *entities.PasswordHistory) error {
	if ph.ChangedAt.IsZero() {
		ph.ChangedAt = time.Now().UTC()
	}
	const q = `
		INSERT INTO password_history (user_id, password_hash, changed_at, changed_by)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := r.db.ExecContext(ctx, q, ph.UserID, ph.PasswordHash, ph.ChangedAt, ph.ChangedBy); err != nil {
		return fmt.Errorf("password_history create: %w", err)
	}
	return nil
}

// GetUserPasswordHistory returns the N most recent password history entries for a user,
// ordered by changed_at DESC.
func (r *PasswordHistoryRepository) GetUserPasswordHistory(
	ctx context.Context, userID string, limit int,
) ([]entities.PasswordHistory, error) {
	const q = `
		SELECT id, user_id, password_hash, changed_at, changed_by
		FROM password_history
		WHERE user_id = $1
		ORDER BY changed_at DESC
		LIMIT $2
	`
	var rows []entities.PasswordHistory
	if err := r.db.SelectContext(ctx, &rows, q, userID, limit); err != nil {
		return nil, fmt.Errorf("password_history get history: %w", err)
	}
	return rows, nil
}

// CheckPasswordInHistory always returns false because bcrypt hashes cannot be compared
// at the SQL level. Hash verification must be performed in the use case layer via
// PasswordHasher.VerifyPassword over the slice returned by GetUserPasswordHistory.
func (r *PasswordHistoryRepository) CheckPasswordInHistory(
	_ context.Context, _, _ string,
) (bool, error) {
	return false, nil
}

// CleanOldPasswordHistory removes password history entries that exceed the keepCount
// most recent entries for the given user.
func (r *PasswordHistoryRepository) CleanOldPasswordHistory(
	ctx context.Context, userID string, keepCount int,
) error {
	const q = `
		DELETE FROM password_history
		WHERE user_id = $1
		  AND id NOT IN (
			  SELECT id FROM password_history
			  WHERE user_id = $1
			  ORDER BY changed_at DESC
			  LIMIT $2
		  )
	`
	if _, err := r.db.ExecContext(ctx, q, userID, keepCount); err != nil {
		return fmt.Errorf("password_history clean: %w", err)
	}
	return nil
}
