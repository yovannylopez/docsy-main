package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// SystemConfigRepository is the sqlx-based implementation for system configuration operations.
type SystemConfigRepository struct {
	db *sqlx.DB
}

// NewSystemConfigRepository creates a new instance of SystemConfigRepository.
func NewSystemConfigRepository(db *sqlx.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// GetConfig retrieves a system configuration entry by key.
// Returns nil, nil when the key does not exist.
func (r *SystemConfigRepository) GetConfig(ctx context.Context, key string) (*entities.SystemConfig, error) {
	const q = `
		SELECT id, key, value, description, is_sensitive, created_at, updated_at, updated_by
		FROM system_config
		WHERE key = $1
	`
	var cfg entities.SystemConfig
	if err := r.db.GetContext(ctx, &cfg, q, key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("system_config get: %w", err)
	}
	return &cfg, nil
}

// SetConfig inserts or updates a system configuration entry.
func (r *SystemConfigRepository) SetConfig(ctx context.Context, config *entities.SystemConfig) error {
	now := time.Now().UTC()
	const q = `
		INSERT INTO system_config (key, value, description, is_sensitive, created_at, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key) DO UPDATE
		SET value       = EXCLUDED.value,
		    description = EXCLUDED.description,
		    is_sensitive = EXCLUDED.is_sensitive,
		    updated_at  = EXCLUDED.updated_at,
		    updated_by  = EXCLUDED.updated_by
	`
	if _, err := r.db.ExecContext(ctx, q,
		config.Key, config.Value, config.Description, config.IsSensitive,
		now, now, config.UpdatedBy,
	); err != nil {
		return fmt.Errorf("system_config set: %w", err)
	}
	return nil
}

// GetAllConfig retrieves all system configuration entries.
func (r *SystemConfigRepository) GetAllConfig(ctx context.Context) ([]entities.SystemConfig, error) {
	const q = `
		SELECT id, key, value, description, is_sensitive, created_at, updated_at, updated_by
		FROM system_config
		ORDER BY key
	`
	var rows []entities.SystemConfig
	if err := r.db.SelectContext(ctx, &rows, q); err != nil {
		return nil, fmt.Errorf("system_config get all: %w", err)
	}
	return rows, nil
}

// DeleteConfig removes a system configuration entry by key.
func (r *SystemConfigRepository) DeleteConfig(ctx context.Context, key string) error {
	const q = `DELETE FROM system_config WHERE key = $1`
	if _, err := r.db.ExecContext(ctx, q, key); err != nil {
		return fmt.Errorf("system_config delete: %w", err)
	}
	return nil
}
