package security

import (
	"context"
	"fmt"

	"github.com/pquerna/otp/totp"
)

// TOTPProviderAdapter wraps github.com/pquerna/otp and implements ports.TOTPProvider.
type TOTPProviderAdapter struct {
	issuer string
}

// NewTOTPProviderAdapter creates a new TOTPProviderAdapter.
// issuer is the application name shown in authenticator apps (e.g. "MyApp").
func NewTOTPProviderAdapter(issuer string) *TOTPProviderAdapter {
	if issuer == "" {
		issuer = "docsy-main"
	}
	return &TOTPProviderAdapter{issuer: issuer}
}

// GenerateSecret creates a new TOTP key and returns the base32 secret and otpauth:// URL.
// The issuer parameter on the method is unused; the adapter uses its own configured issuer.
func (p *TOTPProviderAdapter) GenerateSecret(
	_ context.Context, _, accountName string,
) (secret, qrURL string, err error) {
	key, kerr := totp.Generate(totp.GenerateOpts{
		Issuer:      p.issuer,
		AccountName: accountName,
	})
	if kerr != nil {
		return "", "", fmt.Errorf("totp generate: %w", kerr)
	}
	return key.Secret(), key.URL(), nil
}

// ValidateCode validates a 6-digit TOTP code against the given base32 secret.
// A ±1-period (30 s) clock-skew window is enabled via pquerna/otp defaults.
func (p *TOTPProviderAdapter) ValidateCode(_ context.Context, secret, code string) (bool, error) {
	return totp.Validate(code, secret), nil
}
