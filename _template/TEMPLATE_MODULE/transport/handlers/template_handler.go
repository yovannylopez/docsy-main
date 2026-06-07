// Package handlers contains the Echo HTTP handlers for TEMPLATE_MODULE.
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// TemplateHandler handles HTTP requests for the TEMPLATE_MODULE module.
type TemplateHandler struct {
	createUC ports.CreateTemplateService
	getUC    ports.GetTemplateService
	listUC   ports.ListTemplateService
	updateUC ports.UpdateTemplateService
	deleteUC ports.DeleteTemplateService
}

// NewTemplateHandler creates a new TemplateHandler with its use case dependencies.
func NewTemplateHandler(
	createUC ports.CreateTemplateService,
	getUC ports.GetTemplateService,
	listUC ports.ListTemplateService,
	updateUC ports.UpdateTemplateService,
	deleteUC ports.DeleteTemplateService,
) *TemplateHandler {
	return &TemplateHandler{
		createUC: createUC,
		getUC:    getUC,
		listUC:   listUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

// Create handles POST /api/TEMPLATE_MODULE
func (h *TemplateHandler) Create(c echo.Context) error {
	var req dtos.CreateTemplateRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}

	if req.Name == "" {
		return responses.BadRequest(c, "the 'name' field is required")
	}

	entity, err := h.createUC.Execute(c.Request().Context(), &req)
	if err != nil {
		return responses.MapDomainError(c, err)
	}

	return responses.Created(c, entity, "resource created successfully")
}

// GetByID handles GET /api/TEMPLATE_MODULE/:id
func (h *TemplateHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "the 'id' parameter is required")
	}

	entity, err := h.getUC.Execute(c.Request().Context(), id)
	if err != nil {
		return responses.MapDomainError(c, err)
	}

	return responses.OK(c, entity, "")
}

// List handles GET /api/TEMPLATE_MODULE
func (h *TemplateHandler) List(c echo.Context) error {
	parser := pagination.NewDefaultParser()

	params, err := parser.ParseFromQuery(
		c.QueryParam("limit"),
		c.QueryParam("offset"),
	)
	if err != nil {
		return responses.BadRequest(c, err.Error())
	}

	items, total, err := h.listUC.Execute(c.Request().Context(), params)
	if err != nil {
		return responses.MapDomainError(c, err)
	}

	paginatedResp := pagination.CreateResponse(items, params, total)

	return responses.OKPaginated(c, "", paginatedResp)
}

// Update handles PUT /api/TEMPLATE_MODULE/:id
func (h *TemplateHandler) Update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "the 'id' parameter is required")
	}

	var req dtos.UpdateTemplateRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "invalid request body")
	}

	entity, err := h.updateUC.Execute(c.Request().Context(), id, &req)
	if err != nil {
		return responses.MapDomainError(c, err)
	}

	return responses.OK(c, entity, "resource updated successfully")
}

// Delete handles DELETE /api/TEMPLATE_MODULE/:id
func (h *TemplateHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "the 'id' parameter is required")
	}

	if err := h.deleteUC.Execute(c.Request().Context(), id); err != nil {
		return responses.MapDomainError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
