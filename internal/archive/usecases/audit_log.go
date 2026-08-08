package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	authdomain "github.com/yovannylopez/docsy-main/internal/auth/domain"
	authentities "github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

const (
	auditResourceDocument        = "document"
	auditResourceDocumentFile    = "document_file"
	auditResourceWorkspace       = "workspace"
	auditResourceWorkspaceMember = "workspace_member"
)

func logArchiveAction(
	ctx context.Context,
	auditRepo authports.AuditRepository,
	userID, action, resource, resourceID, message string,
) {
	if auditRepo == nil {
		return
	}
	uid := userID
	res := resource
	rid := resourceID
	msg := message
	log := &authentities.AuditLog{
		ID:         uuid.NewString(),
		UserID:     &uid,
		Action:     action,
		Resource:   &res,
		ResourceID: &rid,
		Result:     authdomain.AuditResultSuccess,
		Message:    &msg,
		CreatedAt:  time.Now().UTC(),
	}
	_ = auditRepo.LogAction(ctx, log)
}
