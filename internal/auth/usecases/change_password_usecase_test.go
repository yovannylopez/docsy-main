package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

// newChangePasswordUseCase builds a fully wired ChangePasswordUseCase for testing.
// Pass nil for passwordHistoryRepo or systemConfigRepo to skip history enforcement.
func newChangePasswordUseCase(
	t *testing.T,
	userRepo *mocks.UserRepository,
	sessionRepo *mocks.SessionRepository,
	hasher *mocks.PasswordHasher,
	auditRepo *mocks.AuditRepository,
	historyRepo *mocks.PasswordHistoryRepository,
	cfgRepo *mocks.SystemConfigRepository,
) *ChangePasswordUseCase {
	t.Helper()
	policy := policies.PasswordPolicy{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSymbol:    true,
		AllowedSymbols:   policies.DefaultSymbols,
	}
	var phRepo ports.PasswordHistoryRepository
	if historyRepo != nil {
		phRepo = historyRepo
	}
	var scRepo ports.SystemConfigRepository
	if cfgRepo != nil {
		scRepo = cfgRepo
	}
	return NewChangePasswordUseCase(userRepo, sessionRepo, hasher, auditRepo, policy, phRepo, scRepo)
}

func newValidUser(userID string) *entities.User {
	stubs := authtest.NewAuthStubs()
	u := authtest.CloneUser(stubs.Entities.ValidUser)
	u.ID = userID
	u.Email = "user@example.com"
	u.PasswordHash = "hashed_current"
	return u
}

// passwordHistoryRecordedBy matches a history row archived on self-service change-password.
func passwordHistoryRecordedBy(userID string) any {
	return mock.MatchedBy(func(ph *entities.PasswordHistory) bool {
		return ph != nil &&
			ph.UserID == userID &&
			ph.ChangedBy != nil &&
			*ph.ChangedBy == userID &&
			ph.PasswordHash != ""
	})
}

// ---------- happy path ----------

func TestChangePasswordUseCase_Execute_Success_NoHistory(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, nil, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "NewPass2@", "hashed_current").Return(false, nil)
	hasher.On("HashPassword", "NewPass2@").Return("hashed_new", nil)
	userRepo.On("UpdatePassword", ctx, userID, "hashed_new").Return(nil)
	userRepo.On("SetMustChangePassword", ctx, userID, false).Return(nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, userID, "password_change").Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	require.NoError(t, uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!",
		NewPassword:     "NewPass2@",
	}))
}

func TestChangePasswordUseCase_Execute_Success_WithHistory(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)
	historyRepo := mocks.NewPasswordHistoryRepository(t)
	cfgRepo := mocks.NewSystemConfigRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, historyRepo, cfgRepo)

	cfgVal := "5"
	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: cfgVal}, nil)

	oldHistory := []entities.PasswordHistory{
		{ID: "h1", UserID: userID, PasswordHash: "hash-old-1"},
	}
	historyRepo.On("GetUserPasswordHistory", ctx, userID, 5).Return(oldHistory, nil)
	// new password does NOT match history
	hasher.On("VerifyPassword", "NewPass2@", "hash-old-1").Return(false, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "NewPass2@", "hashed_current").Return(false, nil)
	hasher.On("HashPassword", "NewPass2@").Return("hashed_new", nil)

	historyRepo.On("Create", ctx, passwordHistoryRecordedBy(userID)).Return(nil)
	userRepo.On("UpdatePassword", ctx, userID, "hashed_new").Return(nil)
	userRepo.On("SetMustChangePassword", ctx, userID, false).Return(nil)
	historyRepo.On("CleanOldPasswordHistory", ctx, userID, 5).Return(nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, userID, "password_change").Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	require.NoError(t, uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!",
		NewPassword:     "NewPass2@",
	}))
}

// ---------- error: user not found ----------

func TestChangePasswordUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, nil, nil)

	userRepo.On("FindByID", ctx, userID).Return((*entities.User)(nil), nil)

	err := uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "any", NewPassword: "any",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCurrentPasswordInvalid)
}

// ---------- error: wrong current password ----------

func TestChangePasswordUseCase_Execute_WrongCurrentPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, nil, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "WrongPass!", "hashed_current").Return(false, nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	err := uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "WrongPass!", NewPassword: "NewPass2@",
	})
	require.ErrorIs(t, err, domain.ErrCurrentPasswordInvalid)
}

// ---------- error: same password ----------

func TestChangePasswordUseCase_Execute_SamePassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, nil, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil) // same-as-current check

	err := uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!", NewPassword: "OldPass1!",
	})
	require.ErrorIs(t, err, domain.ErrSamePassword)
}

// ---------- error: password in history ----------

func TestChangePasswordUseCase_Execute_PasswordInHistory(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)
	historyRepo := mocks.NewPasswordHistoryRepository(t)
	cfgRepo := mocks.NewSystemConfigRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, historyRepo, cfgRepo)

	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: "5"}, nil)

	history := []entities.PasswordHistory{
		{ID: "h1", UserID: userID, PasswordHash: "hash-previously-used"},
	}
	historyRepo.On("GetUserPasswordHistory", ctx, userID, 5).Return(history, nil)
	hasher.On("VerifyPassword", "PreviousPass1!", "hash-previously-used").Return(true, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "PreviousPass1!", "hashed_current").Return(false, nil)

	err := uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!", NewPassword: "PreviousPass1!",
	})
	require.ErrorIs(t, err, domain.ErrPasswordInHistory)
}

// ---------- error: GetUserPasswordHistory fails ----------

func TestChangePasswordUseCase_Execute_HistoryFetchError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)
	historyRepo := mocks.NewPasswordHistoryRepository(t)
	cfgRepo := mocks.NewSystemConfigRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, historyRepo, cfgRepo)

	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: "5"}, nil)
	historyRepo.On("GetUserPasswordHistory", ctx, userID, 5).
		Return(nil, errors.New("db error"))

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "NewPass2@", "hashed_current").Return(false, nil)

	err := uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!", NewPassword: "NewPass2@",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get password history")
}

// ---------- history_count from system config ----------

func TestChangePasswordUseCase_Execute_SystemConfigMissing_UsesDefault(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)
	historyRepo := mocks.NewPasswordHistoryRepository(t)
	cfgRepo := mocks.NewSystemConfigRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, historyRepo, cfgRepo)

	// config key not found → nil, nil → use default 5
	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return((*entities.SystemConfig)(nil), nil)

	historyRepo.On("GetUserPasswordHistory", ctx, userID, defaultPasswordHistoryCount).
		Return([]entities.PasswordHistory{}, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "NewPass2@", "hashed_current").Return(false, nil)
	hasher.On("HashPassword", "NewPass2@").Return("hashed_new", nil)
	historyRepo.On("Create", ctx, passwordHistoryRecordedBy(userID)).Return(nil)
	userRepo.On("UpdatePassword", ctx, userID, "hashed_new").Return(nil)
	userRepo.On("SetMustChangePassword", ctx, userID, false).Return(nil)
	historyRepo.On("CleanOldPasswordHistory", ctx, userID, defaultPasswordHistoryCount).Return(nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, userID, "password_change").Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	require.NoError(t, uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!", NewPassword: "NewPass2@",
	}))
}

// ---------- history Create fails → operation aborted ----------

func TestChangePasswordUseCase_Execute_HistoryCreateFails_AbortsOperation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)
	historyRepo := mocks.NewPasswordHistoryRepository(t)
	cfgRepo := mocks.NewSystemConfigRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, historyRepo, cfgRepo)

	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: "5"}, nil)
	historyRepo.On("GetUserPasswordHistory", ctx, userID, 5).Return([]entities.PasswordHistory{}, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "NewPass2@", "hashed_current").Return(false, nil)
	hasher.On("HashPassword", "NewPass2@").Return("hashed_new", nil)

	historyRepo.On("Create", ctx, passwordHistoryRecordedBy(userID)).
		Return(errors.New("disk full"))

	// UpdatePassword must NOT be called when history Create fails
	err := uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!", NewPassword: "NewPass2@",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "record password history")
	userRepo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
}

// ---------- CleanOldPasswordHistory failure is best-effort (does not fail the op) ----------

func TestChangePasswordUseCase_Execute_CleanHistoryFails_OperationSucceeds(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)
	historyRepo := mocks.NewPasswordHistoryRepository(t)
	cfgRepo := mocks.NewSystemConfigRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, historyRepo, cfgRepo)

	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: "5"}, nil)
	historyRepo.On("GetUserPasswordHistory", ctx, userID, 5).Return([]entities.PasswordHistory{}, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "NewPass2@", "hashed_current").Return(false, nil)
	hasher.On("HashPassword", "NewPass2@").Return("hashed_new", nil)
	historyRepo.On("Create", ctx, passwordHistoryRecordedBy(userID)).Return(nil)
	userRepo.On("UpdatePassword", ctx, userID, "hashed_new").Return(nil)
	userRepo.On("SetMustChangePassword", ctx, userID, false).Return(nil)
	// CleanOldPasswordHistory returns error → should not propagate
	historyRepo.On("CleanOldPasswordHistory", ctx, userID, 5).Return(errors.New("cleanup failed"))
	sessionRepo.On("RevokeAllUserSessions", ctx, userID, "password_change").Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	require.NoError(t, uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!", NewPassword: "NewPass2@",
	}))
}

// ---------- policy validation ----------

func TestChangePasswordUseCase_Execute_WeakNewPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	hasher := mocks.NewPasswordHasher(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := newChangePasswordUseCase(t, userRepo, sessionRepo, hasher, auditRepo, nil, nil)

	userRepo.On("FindByID", ctx, userID).Return(newValidUser(userID), nil)
	hasher.On("VerifyPassword", "OldPass1!", "hashed_current").Return(true, nil)
	hasher.On("VerifyPassword", "short", "hashed_current").Return(false, nil)

	err := uc.Execute(ctx, userID, &dtos.ChangePasswordRequest{
		CurrentPassword: "OldPass1!", NewPassword: "short",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

// ---------- resolveHistoryCount ----------

func TestChangePasswordUseCase_resolveHistoryCount_NilRepo(t *testing.T) {
	uc := &ChangePasswordUseCase{systemConfigRepo: nil}
	assert.Equal(t, defaultPasswordHistoryCount, uc.resolveHistoryCount(context.Background()))
}

func TestChangePasswordUseCase_resolveHistoryCount_InvalidValue(t *testing.T) {
	ctx := context.Background()
	cfgRepo := mocks.NewSystemConfigRepository(t)
	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: "not-a-number"}, nil)

	uc := &ChangePasswordUseCase{systemConfigRepo: cfgRepo}
	assert.Equal(t, defaultPasswordHistoryCount, uc.resolveHistoryCount(ctx))
}

func TestChangePasswordUseCase_resolveHistoryCount_ZeroValue(t *testing.T) {
	ctx := context.Background()
	cfgRepo := mocks.NewSystemConfigRepository(t)
	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: "0"}, nil)

	uc := &ChangePasswordUseCase{systemConfigRepo: cfgRepo}
	assert.Equal(t, defaultPasswordHistoryCount, uc.resolveHistoryCount(ctx))
}

func TestChangePasswordUseCase_resolveHistoryCount_ValidCustomValue(t *testing.T) {
	ctx := context.Background()
	cfgRepo := mocks.NewSystemConfigRepository(t)
	cfgRepo.On("GetConfig", ctx, "password.history_count").
		Return(&entities.SystemConfig{Key: "password.history_count", Value: "10"}, nil)

	uc := &ChangePasswordUseCase{systemConfigRepo: cfgRepo}
	assert.Equal(t, 10, uc.resolveHistoryCount(ctx))
}
