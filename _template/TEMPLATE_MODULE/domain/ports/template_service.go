// Package ports defines the use case interfaces for the TEMPLATE_MODULE module.
package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/entities"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

// CreateTemplateService is the port for the create use case.
type CreateTemplateService interface {
	Execute(ctx context.Context, req *dtos.CreateTemplateRequest) (*entities.TemplateEntity, error)
}

// GetTemplateService is the port for the get-by-id use case.
type GetTemplateService interface {
	Execute(ctx context.Context, id string) (*entities.TemplateEntity, error)
}

// ListTemplateService is the port for the paginated list use case.
// Returns the items, the total count, and any error.
type ListTemplateService interface {
	Execute(ctx context.Context, params *pagination.Params) ([]*entities.TemplateEntity, int, error)
}

// UpdateTemplateService is the port for the partial-update use case.
type UpdateTemplateService interface {
	Execute(ctx context.Context, id string, req *dtos.UpdateTemplateRequest) (*entities.TemplateEntity, error)
}

// DeleteTemplateService is the port for the delete use case.
type DeleteTemplateService interface {
	Execute(ctx context.Context, id string) error
}
