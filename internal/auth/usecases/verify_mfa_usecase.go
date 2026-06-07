package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// VerifyMFAUseCase implements ports.MFAVerifyService (US2, step 2).
type VerifyMFAUseCase struct {
	userRepo       ports.UserRepository
	sessionRepo    ports.SessionRepository
	tokenRepo      ports.VerificationTokenRepository
	tokenGenerator ports.TokenGenerator
	encryptor      ports.MFASecretEncryptor
	totpProvider   ports.TOTPProvider
	auditRepo      ports.AuditRepository
	sessionPolicy  policies.SessionPolicy
}

// NewVerifyMFAUseCase creates a new VerifyMFAUseCase.
func NewVerifyMFAUseCase(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	tokenRepo ports.VerificationTokenRepository,
	tokenGenerator ports.TokenGenerator,
	encryptor ports.MFASecretEncryptor,
	totpProvider ports.TOTPProvider,
	auditRepo ports.AuditRepository,
	sessionPolicy policies.SessionPolicy,
) *VerifyMFAUseCase {
	return &VerifyMFAUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		tokenRepo:      tokenRepo,
		tokenGenerator: tokenGenerator,
		encryptor:      encryptor,
		totpProvider:   totpProvider,
		auditRepo:      auditRepo,
		sessionPolicy:  sessionPolicy,
	}
}

// Verify validates the mfa_challenge token and the TOTP code, then issues a full session.
func (uc *VerifyMFAUseCase) Verify(ctx context.Context, req *dtos.MFAVerifyRequest) (*dtos.LoginResponse, error) {
	// Look up challenge token
	tokenHash := hashToken(req.ChallengeToken)
	record, err := uc.tokenRepo.FindTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrMFAInvalidToken
	}

	if record.TokenType != domain.VerificationTokenTypeMFAChallenge ||
		record.UsedAt != nil ||
		record.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrMFAInvalidToken
	}

	// Load user
	user, err := uc.userRepo.FindByID(ctx, record.UserID)
	if err != nil || user == nil || !user.MFAEnabled || user.MFASecret == nil {
		return nil, domain.ErrMFAInvalidToken
	}

	// Decrypt the stored secret
	plainSecret, err := uc.encryptor.Decrypt(ctx, *user.MFASecret)
	if err != nil {
		return nil, fmt.Errorf("verify mfa: decrypt secret: %w", err)
	}

	// Validate TOTP code
	valid, err := uc.totpProvider.ValidateCode(ctx, plainSecret, req.TOTPCode)
	if err != nil || !valid {
		return nil, domain.ErrMFAInvalidCode
	}

	// Mark challenge token as used (single-use)
	if err := uc.tokenRepo.MarkTokenAsUsed(ctx, record.ID); err != nil {
		return nil, fmt.Errorf("verify mfa: mark token used: %w", err)
	}

	// Delegate to the shared session-creation helper (reused across login and MFA verify)
	loginUC := &LoginUseCase{
		userRepo:       uc.userRepo,
		sessionRepo:    uc.sessionRepo,
		tokenGenerator: uc.tokenGenerator,
		auditRepo:      uc.auditRepo,
		sessionPolicy:  uc.sessionPolicy,
	}
	return loginUC.createSessionAndLoginResponse(ctx, user, user.Email, "", "")
}
