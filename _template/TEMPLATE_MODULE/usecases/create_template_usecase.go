// Package usecases contains the application business logic for TEMPLATE_MODULE.
package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/ports"
)

// CreateTemplateUseCase handles the creation of a new TemplateEntity.
type CreateTemplateUseCase struct {
	repo ports.TemplateRepository
}

// NewCreateTemplateUseCase creates a new instance of CreateTemplateUseCase.
func NewCreateTemplateUseCase(repo ports.TemplateRepository) *CreateTemplateUseCase {
	return &CreateTemplateUseCase{repo: repo}
}

// Execute validates the request and persists a new entity.
func (uc *CreateTemplateUseCase) Execute(ctx context.Context, req *dtos.CreateTemplateRequest) (*entities.TemplateEntity, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	now := time.Now().UTC()
	entity := &entities.TemplateEntity{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to create template entity: %w", err)
	}

	return entity, nil
}
