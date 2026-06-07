package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/security"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

func TestAuthUseCase_Logout_Success_Audits(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessRepo := mocks.NewSessionRepository(t)
	auditRepo := mocks.NewAuditRepository(t)
	tg := security.NewTokenGenerator("unit-test-secret-key-32bytes!!")

	stubs := authtest.NewAuthStubs()
	tokUser := authtest.CloneUser(stubs.Entities.ValidUser)
	tokUser.ID = "u1"
	tokUser.Email = "user@example.com"
	sessionID := "sess-logout-1"
	token, err := tg.GenerateToken(tokUser, sessionID)
	require.NoError(t, err)

	dbUser := authtest.CloneUser(tokUser)
	userRepo.On("FindByID", ctx, "u1").Return(dbUser, nil)
	sessRepo.On("RevokeSession", ctx, sessionID, domain.SessionRevokeReasonLogout).Return(nil)
	auditRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionUserLogout &&
			log.Result == domain.AuditResultSuccess &&
			log.UserID != nil && *log.UserID == "u1" &&
			log.SessionID != nil && *log.SessionID == sessionID &&
			log.Message != nil && *log.Message == "user@example.com" &&
			log.IPAddress != nil && *log.IPAddress == "192.168.1.1" &&
			log.UserAgent != nil && *log.UserAgent == "TestAgent/1.0"
	})).Return(nil)

	uc := NewAuthUseCase(userRepo, tg, sessRepo, auditRepo)
	require.NoError(t, uc.Logout(ctx, token.AccessToken, "TestAgent/1.0", "192.168.1.1"))
}

func TestAuthUseCase_Logout_AuditFailureStillSucceeds(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessRepo := mocks.NewSessionRepository(t)
	auditRepo := mocks.NewAuditRepository(t)
	tg := security.NewTokenGenerator("unit-test-secret-key-32bytes!!")

	stubs := authtest.NewAuthStubs()
	tokUser := authtest.CloneUser(stubs.Entities.ValidUser)
	tokUser.ID = "u1"
	tokUser.Email = "user@example.com"
	sessionID := "sess-logout-2"
	token, err := tg.GenerateToken(tokUser, sessionID)
	require.NoError(t, err)

	userRepo.On("FindByID", ctx, "u1").Return(tokUser, nil)
	sessRepo.On("RevokeSession", ctx, sessionID, domain.SessionRevokeReasonLogout).Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(errors.New("audit db down"))

	uc := NewAuthUseCase(userRepo, tg, sessRepo, auditRepo)
	require.NoError(t, uc.Logout(ctx, token.AccessToken, "", ""))
}

func TestAuthUseCase_Logout_InvalidToken_AuditsFailure(t *testing.T) {
	ctx := context.Background()
	auditRepo := mocks.NewAuditRepository(t)
	tg := security.NewTokenGenerator("unit-test-secret-key-32bytes!!")

	auditRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionUserLogout && log.Result == domain.AuditResultFailure
	})).Return(nil)

	uc := NewAuthUseCase(mocks.NewUserRepository(t), tg, mocks.NewSessionRepository(t), auditRepo)
	err := uc.Logout(ctx, "not-a-jwt", "ua", "1.2.3.4")
	assert.Error(t, err)
}

func TestAuthUseCase_Logout_RevokeFailure_AuditsFailure(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessRepo := mocks.NewSessionRepository(t)
	auditRepo := mocks.NewAuditRepository(t)
	tg := security.NewTokenGenerator("unit-test-secret-key-32bytes!!")

	stubs := authtest.NewAuthStubs()
	tokUser := authtest.CloneUser(stubs.Entities.ValidUser)
	tokUser.ID = "u1"
	tokUser.PasswordChangedAt = time.Now()
	sessionID := "sess-logout-3"
	token, err := tg.GenerateToken(tokUser, sessionID)
	require.NoError(t, err)

	userRepo.On("FindByID", ctx, "u1").Return(tokUser, nil)
	sessRepo.On("RevokeSession", ctx, sessionID, domain.SessionRevokeReasonLogout).Return(errors.New("db error"))
	auditRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Result == domain.AuditResultFailure
	})).Return(nil)

	uc := NewAuthUseCase(userRepo, tg, sessRepo, auditRepo)
	err = uc.Logout(ctx, token.AccessToken, "", "")
	assert.Error(t, err)
}
