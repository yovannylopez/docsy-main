package identity

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
)

// LDAPProvider is a LDAP/AD configuration container; full integration is left as an extension.
type LDAPProvider struct {
	cfg config.LDAPConfig
}

// NewLDAPProvider creates the provider based on configuration.
func NewLDAPProvider(cfg config.LDAPConfig) *LDAPProvider {
	return &LDAPProvider{cfg: cfg}
}

// Enabled reflects AUTH_LDAP_ENABLED.
func (p *LDAPProvider) Enabled() bool {
	return p.cfg.Enabled
}

// Authenticate reserved: synchronization with local DB and LDAP bind not yet implemented.
func (p *LDAPProvider) Authenticate(_ context.Context, _, _ string) (*entities.User, error) {
	if !p.cfg.Enabled {
		return nil, nil
	}
	return nil, fmt.Errorf("ldap: active directory integration not implemented")
}
