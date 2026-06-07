package usecases

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// DisableMFAUseCase implements ports.MFADisableService (US3).
type DisableMFAUseCase struct {
	userRepo  ports.UserRepository
	totp      ports.TOTPProvider
	encryptor ports.MFASecretEncryptor
}

// NewDisableMFAUseCase creates a new DisableMFAUseCase.
func NewDisableMFAUseCase(
	userRepo ports.UserRepository,
	totp ports.TOTPProvider,
	encryptor ports.MFASecretEncryptor,
) *DisableMFAUseCase {
	return &DisableMFAUseCase{
		userRepo:  userRepo,
		totp:      totp,
		encryptor: encryptor,
	}
}

// Disable validates the TOTP code and then disables MFA for the user.
func (uc *DisableMFAUseCase) Disable(ctx context.Context, userID string, req *dtos.MFADisableRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("disable mfa: user not found: %w", domain.ErrMFANotEnabled)
	}

	if !user.MFAEnabled || user.MFASecret == nil {
		return domain.ErrMFANotEnabled
	}

	// Decrypt the stored secret
	plainSecret, err := uc.encryptor.Decrypt(ctx, *user.MFASecret)
	if err != nil {
		return fmt.Errorf("disable mfa: decrypt secret: %w", err)
	}

	// Validate the supplied TOTP code
	valid, err := uc.totp.ValidateCode(ctx, plainSecret, req.TOTPCode)
	if err != nil || !valid {
		return domain.ErrMFAInvalidCode
	}

	// Disable MFA and clear the secret
	if err := uc.userRepo.DisableMFA(ctx, userID); err != nil {
		return fmt.Errorf("disable mfa: disable: %w", err)
	}

	return nil
}
