package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
)

// MFASetupService defines the use case for initiating MFA TOTP setup (US1, step 1).
type MFASetupService interface {
	// Setup generates a new TOTP secret for the authenticated user and returns
	// the QR URL, the plain secret, and a single-use mfa_setup token.
	Setup(ctx context.Context, userID string) (*dtos.MFASetupResponse, error)
}

// MFAConfirmService defines the use case for confirming MFA TOTP setup (US1, step 2).
type MFAConfirmService interface {
	// Confirm validates the setup token and the TOTP code, then enables MFA for the user.
	Confirm(ctx context.Context, userID string, req *dtos.MFAConfirmRequest) error
}

// MFAVerifyService defines the use case for verifying an MFA challenge during login (US2).
type MFAVerifyService interface {
	// Verify validates the challenge token and the TOTP code, then issues a full session.
	Verify(ctx context.Context, req *dtos.MFAVerifyRequest) (*dtos.LoginResponse, error)
}

// MFADisableService defines the use case for disabling MFA (US3).
type MFADisableService interface {
	// Disable validates the TOTP code and then disables MFA for the user.
	Disable(ctx context.Context, userID string, req *dtos.MFADisableRequest) error
}
