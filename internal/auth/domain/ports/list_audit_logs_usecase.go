package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// ListAuditLogsUseCase defines the interface for the list audit logs use case
type ListAuditLogsUseCase interface {
	Execute(ctx context.Context, filters *dtos.AuditLogFilters) ([]entities.AuditLog, int, error)
}
