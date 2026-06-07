package usecases

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/ports"
)

// GetTemplateUseCase retrieves a single TemplateEntity by its ID.
type GetTemplateUseCase struct {
	repo ports.TemplateRepository
}

// NewGetTemplateUseCase creates a new GetTemplateUseCase.
func NewGetTemplateUseCase(repo ports.TemplateRepository) *GetTemplateUseCase {
	return &GetTemplateUseCase{repo: repo}
}

// Execute returns the entity identified by id, or an error if it does not exist.
func (uc *GetTemplateUseCase) Execute(ctx context.Context, id string) (*entities.TemplateEntity, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template entity: %w", err)
	}

	return entity, nil
}
