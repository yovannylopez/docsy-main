// Package repositories contains the database adapters for TEMPLATE_MODULE.
package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/entities"
	pkgerrors "github.com/yovannylopez/docsy-main/pkg/errors"
)

// TemplateRepositoryAdapter implements ports.TemplateRepository using sqlx + PostgreSQL.
type TemplateRepositoryAdapter struct {
	db *sqlx.DB
}

// NewTemplateRepositoryAdapter creates a new TemplateRepositoryAdapter.
func NewTemplateRepositoryAdapter(db *sqlx.DB) *TemplateRepositoryAdapter {
	return &TemplateRepositoryAdapter{db: db}
}

// Create inserts a new entity into the database.
func (r *TemplateRepositoryAdapter) Create(ctx context.Context, e *entities.TemplateEntity) error {
	const query = `
		INSERT INTO template_entities (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		e.ID, e.Name, e.Description, e.IsActive, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert template entity: %w", err)
	}

	return nil
}

// GetByID retrieves a single entity by its primary key.
func (r *TemplateRepositoryAdapter) GetByID(ctx context.Context, id string) (*entities.TemplateEntity, error) {
	var e entities.TemplateEntity
	const query = `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM template_entities
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &e, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get template entity %s: %w", id, err)
	}

	return &e, nil
}

// List returns a paginated slice of entities.
func (r *TemplateRepositoryAdapter) List(ctx context.Context, limit, offset int) ([]*entities.TemplateEntity, error) {
	var items []*entities.TemplateEntity
	const query = `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM template_entities
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	if err := r.db.SelectContext(ctx, &items, query, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list template entities: %w", err)
	}

	return items, nil
}

// Count returns the total number of entities.
func (r *TemplateRepositoryAdapter) Count(ctx context.Context) (int, error) {
	var count int
	const query = `SELECT COUNT(*) FROM template_entities`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, fmt.Errorf("failed to count template entities: %w", err)
	}

	return count, nil
}

// Update persists changes to an existing entity.
func (r *TemplateRepositoryAdapter) Update(ctx context.Context, e *entities.TemplateEntity) error {
	const query = `
		UPDATE template_entities
		SET name = $2, description = $3, is_active = $4, updated_at = $5
		WHERE id = $1
	`
	e.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		e.ID, e.Name, e.Description, e.IsActive, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update template entity %s: %w", e.ID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}

// Delete removes an entity by id.
func (r *TemplateRepositoryAdapter) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM template_entities WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template entity %s: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}
