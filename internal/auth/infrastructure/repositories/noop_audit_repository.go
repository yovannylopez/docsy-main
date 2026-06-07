package repositories

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// NoOpAuditRepository is a no-op implementation that does nothing for the audit repository
type NoOpAuditRepository struct{}

// NewNoOpAuditRepository creates a new instance of the no-op implementation
func NewNoOpAuditRepository() ports.AuditRepository {
	return &NoOpAuditRepository{}
}

// LogAction does nothing (no-op)
func (r *NoOpAuditRepository) LogAction(ctx context.Context, log *entities.AuditLog) error {
	return nil
}

// GetUserAuditLogs does nothing (no-op)
func (r *NoOpAuditRepository) GetUserAuditLogs(ctx context.Context, userID string, limit, offset int) ([]entities.AuditLog, error) {
	return []entities.AuditLog{}, nil
}

// GetSessionAuditLogs does nothing (no-op)
func (r *NoOpAuditRepository) GetSessionAuditLogs(ctx context.Context, sessionID string, limit, offset int) ([]entities.AuditLog, error) {
	return []entities.AuditLog{}, nil
}

// List returns an empty list and total 0
func (r *NoOpAuditRepository) List(ctx context.Context, filters *dtos.AuditLogFilters) ([]entities.AuditLog, int, error) {
	return []entities.AuditLog{}, 0, nil
}
