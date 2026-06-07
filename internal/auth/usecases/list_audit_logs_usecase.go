package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// Ensure ListAuditLogsUseCase implements ports.ListAuditLogsUseCase
var _ ports.ListAuditLogsUseCase = (*ListAuditLogsUseCase)(nil)

// ListAuditLogsUseCase handles the business logic for listing audit logs
type ListAuditLogsUseCase struct {
	auditRepo ports.AuditRepository
}

// NewListAuditLogsUseCase creates a new ListAuditLogsUseCase instance
func NewListAuditLogsUseCase(auditRepo ports.AuditRepository) *ListAuditLogsUseCase {
	return &ListAuditLogsUseCase{
		auditRepo: auditRepo,
	}
}

// Execute runs the audit log listing use case
func (uc *ListAuditLogsUseCase) Execute(ctx context.Context, filters *dtos.AuditLogFilters) ([]entities.AuditLog, int, error) {
	// If filters is nil, return an error
	if filters == nil {
		return nil, 0, fmt.Errorf("filters cannot be nil")
	}

	// Validate filters before applying defaults
	if err := uc.validateFiltersBeforeDefaults(filters); err != nil {
		return nil, 0, fmt.Errorf("invalid filters: %w", err)
	}

	// Apply default values
	uc.applyDefaults(filters)

	// Validate filters after applying defaults
	if err := uc.validateFilters(filters); err != nil {
		return nil, 0, fmt.Errorf("invalid filters: %w", err)
	}

	// Retrieve logs from the repository
	logs, total, err := uc.auditRepo.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return logs, total, nil
}

// validateFiltersBeforeDefaults validates filters that must not have invalid values before applying defaults
func (uc *ListAuditLogsUseCase) validateFiltersBeforeDefaults(filters *dtos.AuditLogFilters) error {
	// Validate negative offset
	if filters.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	// Validate limit if specified (0 is valid and the default will be applied)
	if filters.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}

	if filters.Limit > constants.MaxPageSize {
		return fmt.Errorf("limit cannot be greater than %d", constants.MaxPageSize)
	}

	// Validate date range
	if filters.StartDate != nil && filters.EndDate != nil {
		if filters.StartDate.After(*filters.EndDate) {
			return fmt.Errorf("start_date cannot be after end_date")
		}
	}

	// Validate enum values for result
	if filters.Result != nil && !domain.ValidAuditResult(*filters.Result) {
		return fmt.Errorf("invalid result value: %s (must be %s, %s, or %s)",
			*filters.Result, domain.AuditResultSuccess, domain.AuditResultFailure, domain.AuditResultError)
	}

	return nil
}

// validateFilters validates search filters after applying defaults
func (uc *ListAuditLogsUseCase) validateFilters(filters *dtos.AuditLogFilters) error {
	if filters == nil {
		return fmt.Errorf("filters cannot be nil")
	}

	// Validate pagination limits
	if filters.Limit < 1 {
		return fmt.Errorf("limit must be greater than 0")
	}

	if filters.Limit > constants.MaxPageSize {
		return fmt.Errorf("limit cannot be greater than %d", constants.MaxPageSize)
	}

	if filters.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	// Validate date range
	if filters.StartDate != nil && filters.EndDate != nil {
		if filters.StartDate.After(*filters.EndDate) {
			return fmt.Errorf("start_date cannot be after end_date")
		}
	}

	// Validate enum values for result
	if filters.Result != nil && !domain.ValidAuditResult(*filters.Result) {
		return fmt.Errorf("invalid result value: %s", *filters.Result)
	}

	return nil
}

// applyDefaults applies default values to the filters
func (uc *ListAuditLogsUseCase) applyDefaults(filters *dtos.AuditLogFilters) {
	if filters.Limit == 0 {
		filters.Limit = 20
	}

	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// If no start date is set, default to 30 days ago
	if filters.StartDate == nil && filters.EndDate == nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		filters.StartDate = &thirtyDaysAgo
	}
}
