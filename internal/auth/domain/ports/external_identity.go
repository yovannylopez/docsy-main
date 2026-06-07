package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
)

// ExternalIdentityProvider autentica contra LDAP/Active Directory u otros directorios.
// Cuando Enabled() es false, LoginUseCase solo usa credenciales locales.
type ExternalIdentityProvider interface {
	Enabled() bool
	Authenticate(ctx context.Context, email, password string) (*entities.User, error)
}
