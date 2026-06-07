//go:build ignore

// Illustrative example (not part of the compilable module; copy to the handlers slice if applicable).

package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// Example of a refactored handler
type ExampleHandler struct {
	// ... other fields
}

// List handles GET /api/v1/example
func (h *ExampleHandler) List(c echo.Context) error {
	// Create pagination parser
	parser := pagination.NewDefaultParser()

	// Parse pagination parameters
	params, err := parser.ParseFromQuery(
		c.QueryParam("limit"),
		c.QueryParam("offset"),
	)
	if err != nil {
		return responses.BadRequest(c, err.Error())
	}

	// Execute use case with parsed parameters
	data, total, err := h.usecase.Execute(c.Request().Context(), params.Limit, params.Offset)
	if err != nil {
		return responses.InternalError(c, err.Error())
	}

	page := pagination.CreateResponse(data, params, total)
	return responses.OKPaginated(c, "data retrieved successfully", page)
}

// Comparison: Code BEFORE vs AFTER

// BEFORE (duplicated code in each handler):
/*
func (h *MyHandler) List(c echo.Context) error {
	// Get pagination parameters
	limit := 10 // default value
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0 // default value
	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil {
			offset = parsedOffset
		}
	}

	// Validate parameters (conceptual equivalent of pkg/pagination.Validate)
	if limit < 1 {
		return responses.BadRequest(c, "pagination: limit out of allowed range: minimum 1")
	}
	if limit > 100 {
		return responses.BadRequest(c, "pagination: limit out of allowed range: maximum 100")
	}
	if offset < 0 {
		return responses.BadRequest(c, "pagination: negative offset")
	}

	// Execute use case
	data, total, err := h.usecase.Execute(c.Request().Context(), limit, offset)
	if err != nil {
		return responses.InternalError(c, err.Error())
	}

	// Manually build pagination metadata
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	currentPage := (offset / limit) + 1

	response := map[string]any{
		"data": data,
		"pagination": map[string]any{
			"total":        total,
			"limit":        limit,
			"offset":       offset,
			"total_pages":  totalPages,
			"current_page": currentPage,
			"has_next":     currentPage < totalPages,
			"has_previous": currentPage > 1,
		},
	}

	return responses.OK(c, response, "data retrieved successfully")
}
*/

// AFTER (centralized and clean code):
/*
func (h *MyHandler) List(c echo.Context) error {
	// Create pagination parser
	parser := pagination.NewDefaultParser()

	// Parse pagination parameters
	params, err := parser.ParseFromQuery(
		c.QueryParam("limit"),
		c.QueryParam("offset"),
	)
	if err != nil {
		return responses.BadRequest(c, err.Error())
	}

	// Execute use case with parsed parameters
	data, total, err := h.usecase.Execute(c.Request().Context(), params.Limit, params.Offset)
	if err != nil {
		return responses.InternalError(c, err.Error())
	}

	response := pagination.CreateResponse(data, params, total)
	return responses.OKPaginated(c, "data retrieved successfully", response)
}
*/
