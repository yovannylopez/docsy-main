// Package ports defines the interfaces (ports) for the TEMPLATE_MODULE module.
// These are implemented by the infrastructure layer (adapters).
package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/entities"
)

// TemplateRepository defines the persistence operations for TemplateEntity.
// Replace TemplateEntity with your domain entity.
type TemplateRepository interface {
	// Create persists a new entity. Returns an error if it already exists.
	Create(ctx context.Context, entity *entities.TemplateEntity) error

	// GetByID retrieves an entity by its unique identifier.
	// Returns pkg/errors.ErrNotFound if no entity matches the given id.
	GetByID(ctx context.Context, id string) (*entities.TemplateEntity, error)

	// List returns a paginated slice of entities.
	List(ctx context.Context, limit, offset int) ([]*entities.TemplateEntity, error)

	// Count returns the total number of entities (for pagination metadata).
	Count(ctx context.Context) (int, error)

	// Update persists changes to an existing entity.
	Update(ctx context.Context, entity *entities.TemplateEntity) error

	// Delete removes an entity by id. Soft or hard delete — your choice.
	Delete(ctx context.Context, id string) error
}
