package handlers

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// AuditHandler handles HTTP requests related to auditing
type AuditHandler struct {
	listAuditLogsUC ports.ListAuditLogsUseCase
}

// NewAuditHandler creates a new AuditHandler instance
func NewAuditHandler(listAuditLogsUC ports.ListAuditLogsUseCase) *AuditHandler {
	return &AuditHandler{
		listAuditLogsUC: listAuditLogsUC,
	}
}

// List handles GET /api/v1/audit
func (h *AuditHandler) List(c echo.Context) error {
	parser := pagination.NewDefaultParser()

	// Parse pagination parameters
	params, err := parser.ParseFromQuery(
		c.QueryParam("limit"),
		c.QueryParam("offset"),
	)
	if err != nil {
		return responses.BadRequest(c, err.Error())
	}

	// Build filters from query parameters
	filters := &dtos.AuditLogFilters{
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	// Parse optional filters
	if userID := c.QueryParam("user_id"); userID != "" {
		filters.UserID = &userID
	}

	if sessionID := c.QueryParam("session_id"); sessionID != "" {
		filters.SessionID = &sessionID
	}

	if action := c.QueryParam("action"); action != "" {
		filters.Action = &action
	}

	if resource := c.QueryParam("resource"); resource != "" {
		filters.Resource = &resource
	}

	if resourceID := c.QueryParam("resource_id"); resourceID != "" {
		filters.ResourceID = &resourceID
	}

	if result := c.QueryParam("result"); result != "" {
		filters.Result = &result
	}

	if message := c.QueryParam("message"); message != "" {
		filters.Message = &message
	}

	// Parse dates
	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return responses.BadRequest(c, fmt.Sprintf("invalid start_date format: %v", err))
		}
		filters.StartDate = &startDate
	}

	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return responses.BadRequest(c, fmt.Sprintf("invalid end_date format: %v", err))
		}
		// Adjust end_date to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filters.EndDate = &endDate
	}

	// Execute use case
	logs, total, err := h.listAuditLogsUC.Execute(c.Request().Context(), filters)
	if err != nil {
		return err
	}

	paginatedResponse := pagination.CreateResponse(logs, params, total)
	return responses.OKPaginated(c, "audit logs retrieved successfully", paginatedResponse)
}
