package identity

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// NoopExternalIdentity no activa directorio externo.
type NoopExternalIdentity struct{}

// Enabled siempre false.
func (NoopExternalIdentity) Enabled() bool {
	return false
}

// Authenticate no debe llamarse si Enabled es false.
func (NoopExternalIdentity) Authenticate(_ context.Context, _, _ string) (*entities.User, error) {
	return nil, nil
}
