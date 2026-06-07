package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// ConfirmMFAUseCase implements ports.MFAConfirmService (US1, step 2).
type ConfirmMFAUseCase struct {
	userRepo  ports.UserRepository
	tokenRepo ports.VerificationTokenRepository
	totp      ports.TOTPProvider
	encryptor ports.MFASecretEncryptor
}

// NewConfirmMFAUseCase creates a new ConfirmMFAUseCase.
func NewConfirmMFAUseCase(
	userRepo ports.UserRepository,
	tokenRepo ports.VerificationTokenRepository,
	totp ports.TOTPProvider,
	encryptor ports.MFASecretEncryptor,
) *ConfirmMFAUseCase {
	return &ConfirmMFAUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		totp:      totp,
		encryptor: encryptor,
	}
}

// Confirm validates the mfa_setup token and the TOTP code, then enables MFA for the user.
func (uc *ConfirmMFAUseCase) Confirm(ctx context.Context, userID string, req *dtos.MFAConfirmRequest) error {
	// Look up the setup token by its hash
	tokenHash := hashToken(req.SetupToken)
	record, err := uc.tokenRepo.FindTokenByHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrMFAInvalidToken
	}

	// Validate token ownership, type, expiry and usage
	if record.UserID != userID ||
		record.TokenType != domain.VerificationTokenTypeMFASetup ||
		record.UsedAt != nil ||
		record.ExpiresAt.Before(time.Now()) {
		return domain.ErrMFAInvalidToken
	}

	// Load the user to get the pending encrypted secret
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil || user.MFASecret == nil {
		return fmt.Errorf("confirm mfa: user or pending secret not found: %w", domain.ErrMFAInvalidToken)
	}

	// Decrypt the pending secret
	plainSecret, err := uc.encryptor.Decrypt(ctx, *user.MFASecret)
	if err != nil {
		return fmt.Errorf("confirm mfa: decrypt secret: %w", err)
	}

	// Validate the TOTP code
	valid, err := uc.totp.ValidateCode(ctx, plainSecret, req.TOTPCode)
	if err != nil || !valid {
		return domain.ErrMFAInvalidCode
	}

	// Mark token as used
	if err := uc.tokenRepo.MarkTokenAsUsed(ctx, record.ID); err != nil {
		return fmt.Errorf("confirm mfa: mark token used: %w", err)
	}

	// Enable MFA — store the encrypted secret with mfa_enabled = true
	if err := uc.userRepo.EnableMFA(ctx, userID, *user.MFASecret); err != nil {
		return fmt.Errorf("confirm mfa: enable mfa: %w", err)
	}

	return nil
}

// hashToken returns the SHA-256 hex digest of a raw token string.
func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
