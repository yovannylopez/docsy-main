package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// AuditRepository defines the audit repository operations
type AuditRepository interface {
	LogAction(ctx context.Context, log *entities.AuditLog) error
	GetUserAuditLogs(ctx context.Context, userID string, limit, offset int) ([]entities.AuditLog, error)
	GetSessionAuditLogs(ctx context.Context, sessionID string, limit, offset int) ([]entities.AuditLog, error)
	// List returns audit logs with advanced filters
	List(ctx context.Context, filters *dtos.AuditLogFilters) ([]entities.AuditLog, int, error)
}
