package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// logAuthEvent persists an audit row; failures are ignored (best-effort).
func logAuthEvent(
	ctx context.Context,
	auditRepo ports.AuditRepository,
	userID, sessionID *string,
	action, result, message, ipAddress, userAgent string,
) {
	if auditRepo == nil {
		return
	}

	requestID := "system-generated"
	log := &entities.AuditLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		SessionID: sessionID,
		Action:    action,
		Result:    result,
		Message:   optionalString(message),
		IPAddress: optionalString(ipAddress),
		UserAgent: optionalString(userAgent),
		RequestID: &requestID,
		CreatedAt: time.Now(),
	}
	_ = auditRepo.LogAction(ctx, log)
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
