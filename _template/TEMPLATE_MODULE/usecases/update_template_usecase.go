package usecases

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/ports"
)

// UpdateTemplateUseCase applies partial updates to an existing TemplateEntity.
type UpdateTemplateUseCase struct {
	repo ports.TemplateRepository
}

// NewUpdateTemplateUseCase creates a new UpdateTemplateUseCase.
func NewUpdateTemplateUseCase(repo ports.TemplateRepository) *UpdateTemplateUseCase {
	return &UpdateTemplateUseCase{repo: repo}
}

// Execute fetches the entity, applies the updates from req, and persists the changes.
func (uc *UpdateTemplateUseCase) Execute(ctx context.Context, id string, req *dtos.UpdateTemplateRequest) (*entities.TemplateEntity, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find template entity for update: %w", err)
	}

	if req.Name != nil {
		entity.Name = *req.Name
	}

	if req.Description != nil {
		entity.Description = req.Description
	}

	if err := uc.repo.Update(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to update template entity: %w", err)
	}

	return entity, nil
}
