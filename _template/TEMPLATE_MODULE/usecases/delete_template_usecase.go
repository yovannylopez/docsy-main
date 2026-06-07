package usecases

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/ports"
)

// DeleteTemplateUseCase removes a TemplateEntity by its ID.
type DeleteTemplateUseCase struct {
	repo ports.TemplateRepository
}

// NewDeleteTemplateUseCase creates a new DeleteTemplateUseCase.
func NewDeleteTemplateUseCase(repo ports.TemplateRepository) *DeleteTemplateUseCase {
	return &DeleteTemplateUseCase{repo: repo}
}

// Execute deletes the entity with the given id.
// Returns pkg/errors.ErrNotFound (wrapped) if no entity matches the id.
func (uc *DeleteTemplateUseCase) Execute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete template entity: %w", err)
	}

	return nil
}
