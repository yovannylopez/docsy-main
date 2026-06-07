// Package dtos defines the Data Transfer Objects for the TEMPLATE_MODULE module.
package dtos

// CreateTemplateRequest is the DTO for creating a new entity.
type CreateTemplateRequest struct {
	Name        string  `json:"name"        validate:"required,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// UpdateTemplateRequest is the DTO for updating an existing entity.
type UpdateTemplateRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// TemplateResponse is the DTO returned by the API for a single entity.
type TemplateResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
