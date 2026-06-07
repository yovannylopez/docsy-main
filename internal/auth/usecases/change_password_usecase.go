package usecases

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
)

// Sentinel errors are now in the domain package (internal/auth/domain/errors.go).
// Re-exported here for backwards compatibility in case test code references them.
var (
	ErrCurrentPasswordInvalid = domain.ErrCurrentPasswordInvalid
	ErrSamePassword           = domain.ErrSamePassword
)

const defaultPasswordHistoryCount = 5

// ChangePasswordUseCase handles password rotation with full session revocation (FR-001 … FR-008)
// and password history enforcement (PH-FR-001 … PH-FR-007).
type ChangePasswordUseCase struct {
	userRepo            ports.UserRepository
	sessionRepo         ports.SessionRepository
	passwordHasher      ports.PasswordHasher
	auditRepo           ports.AuditRepository
	passwordPolicy      policies.PasswordPolicy
	passwordHistoryRepo ports.PasswordHistoryRepository
	systemConfigRepo    ports.SystemConfigRepository
}

// NewChangePasswordUseCase creates a ChangePasswordUseCase with all required dependencies.
// passwordHistoryRepo and systemConfigRepo are optional (nil-safe); when nil, history
// enforcement is skipped.
func NewChangePasswordUseCase(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	passwordHasher ports.PasswordHasher,
	auditRepo ports.AuditRepository,
	passwordPolicy policies.PasswordPolicy,
	passwordHistoryRepo ports.PasswordHistoryRepository,
	systemConfigRepo ports.SystemConfigRepository,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo:            userRepo,
		sessionRepo:         sessionRepo,
		passwordHasher:      passwordHasher,
		auditRepo:           auditRepo,
		passwordPolicy:      passwordPolicy,
		passwordHistoryRepo: passwordHistoryRepo,
		systemConfigRepo:    systemConfigRepo,
	}
}

// Execute changes the authenticated user's password.
//
// Ordering (policy-before-side-effects per FR-003 / US2, history check per PH-FR-002):
//  1. Load user
//  2. Verify current password         → ErrCurrentPasswordInvalid on failure
//  3. Reject new == current           → ErrSamePassword
//  4. Check password history          → ErrPasswordInHistory if reused
//  5. Validate new against policy     → policy error (explicit)
//  6. Hash new password
//  7. Record old hash in history      → error is fatal (before the update)
//  8. Persist new hash + metadata
//  9. Clear must_change_password flag
//
// 10. Clean old history entries       → best-effort (housekeeping only)
// 11. Revoke all sessions
// 12. Audit (best-effort; non-fatal)
func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID string, req *dtos.ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("change password: load user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("change password: %w", ErrCurrentPasswordInvalid)
	}

	// Step 2 — verify current password (generic error; see FR-002).
	valid, err := uc.passwordHasher.VerifyPassword(req.CurrentPassword, user.PasswordHash)
	if err != nil || !valid {
		uc.logAudit(ctx, userID, domain.AuditActionPasswordChangeFailed, domain.AuditResultFailure)
		return ErrCurrentPasswordInvalid
	}

	// Step 3 — reject if same as current.
	sameAsCurrent, err := uc.passwordHasher.VerifyPassword(req.NewPassword, user.PasswordHash)
	if err == nil && sameAsCurrent {
		return ErrSamePassword
	}

	// Step 4 — check against password history (PH-FR-002).
	historyCount := uc.resolveHistoryCount(ctx)
	if uc.passwordHistoryRepo != nil && historyCount > 0 {
		history, err := uc.passwordHistoryRepo.GetUserPasswordHistory(ctx, userID, historyCount)
		if err != nil {
			return fmt.Errorf("change password: get password history: %w", err)
		}
		for _, entry := range history {
			inHistory, verr := uc.passwordHasher.VerifyPassword(req.NewPassword, entry.PasswordHash)
			if verr == nil && inHistory {
				return domain.ErrPasswordInHistory
			}
		}
	}

	// Step 5 — validate new password policy (explicit feedback, FR-003 / US2).
	if err := uc.passwordPolicy.Validate(req.NewPassword); err != nil {
		return err
	}

	// Step 6 — hash new password.
	newHash, err := uc.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("change password: hash new password: %w", err)
	}

	// Step 7 — record the OLD hash in history before overwriting it (PH-FR-001).
	// This step is intentionally not best-effort: if history cannot be recorded,
	// we abort before modifying the stored password to preserve data integrity.
	if uc.passwordHistoryRepo != nil {
		actorID := userID
		ph := &entities.PasswordHistory{
			UserID:       userID,
			PasswordHash: user.PasswordHash,
			ChangedAt:    time.Now().UTC(),
			ChangedBy:    &actorID,
		}
		if err := uc.passwordHistoryRepo.Create(ctx, ph); err != nil {
			return fmt.Errorf("change password: record password history: %w", err)
		}
	}

	// Step 8 — persist new hash and reset must_change_password (FR-004).
	if err := uc.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("change password: update password: %w", err)
	}
	if err := uc.userRepo.SetMustChangePassword(ctx, userID, false); err != nil {
		return fmt.Errorf("change password: clear must_change_password: %w", err)
	}

	// Step 10 — prune entries beyond the configured limit (best-effort housekeeping).
	if uc.passwordHistoryRepo != nil && historyCount > 0 {
		_ = uc.passwordHistoryRepo.CleanOldPasswordHistory(ctx, userID, historyCount)
	}

	// Step 11 — revoke ALL sessions so no refresh token remains valid (FR-005).
	if err := uc.sessionRepo.RevokeAllUserSessions(ctx, userID, "password_change"); err != nil {
		return fmt.Errorf("change password: revoke sessions: %w", err)
	}

	// Step 12 — audit event without secrets (FR-008).  Best-effort: a logging failure
	// must not roll back the password change that already succeeded.
	uc.logAudit(ctx, userID, domain.AuditActionPasswordChanged, domain.AuditResultSuccess)

	return nil
}

// resolveHistoryCount reads password.history_count from system_config.
// Falls back to defaultPasswordHistoryCount if the repo is nil, the key is absent,
// the value cannot be parsed, or the value is non-positive.
func (uc *ChangePasswordUseCase) resolveHistoryCount(ctx context.Context) int {
	if uc.systemConfigRepo == nil {
		return defaultPasswordHistoryCount
	}
	cfg, err := uc.systemConfigRepo.GetConfig(ctx, "password.history_count")
	if err != nil || cfg == nil {
		return defaultPasswordHistoryCount
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n <= 0 {
		return defaultPasswordHistoryCount
	}
	return n
}

// logAudit records a password-change audit event.  Errors are suppressed because
// audit failures must not affect the outcome visible to the user.
func (uc *ChangePasswordUseCase) logAudit(ctx context.Context, userID, action, result string) {
	if uc.auditRepo == nil {
		return
	}
	resource := "user"
	log := &entities.AuditLog{
		UserID:     &userID,
		Action:     action,
		Resource:   &resource,
		ResourceID: &userID,
		Result:     result,
		CreatedAt:  time.Now(),
	}
	_ = uc.auditRepo.LogAction(ctx, log)
}
