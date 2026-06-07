package usecases

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

// ListTemplateUseCase returns a paginated list of TemplateEntity records.
type ListTemplateUseCase struct {
	repo ports.TemplateRepository
}

// NewListTemplateUseCase creates a new ListTemplateUseCase.
func NewListTemplateUseCase(repo ports.TemplateRepository) *ListTemplateUseCase {
	return &ListTemplateUseCase{repo: repo}
}

// Execute returns a slice of entities and the total count for pagination metadata.
func (uc *ListTemplateUseCase) Execute(ctx context.Context, params *pagination.Params) ([]*entities.TemplateEntity, int, error) {
	if params == nil {
		params = &pagination.Params{
			Limit:  pagination.DefaultConfig.DefaultLimit,
			Offset: 0,
		}
	}

	items, err := uc.repo.List(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list template entities: %w", err)
	}

	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count template entities: %w", err)
	}

	return items, total, nil
}
