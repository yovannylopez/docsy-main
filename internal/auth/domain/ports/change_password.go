package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
)

// ChangePasswordService is the port for the change-password use case.
type ChangePasswordService interface {
	Execute(ctx context.Context, userID string, req *dtos.ChangePasswordRequest) error
}
