package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

// buildVerifyMFAUseCase is a helper to create a VerifyMFAUseCase with all mocks wired.
func buildVerifyMFAUseCase(
	t *testing.T,
	userRepo *mocks.UserRepository,
	sessionRepo *mocks.SessionRepository,
	tokenRepo *mocks.VerificationTokenRepository,
	tokenGen *mocks.TokenGenerator,
	encryptor *mocks.MFASecretEncryptor,
	totpProvider *mocks.TOTPProvider,
	auditRepo *mocks.AuditRepository,
) *VerifyMFAUseCase {
	t.Helper()
	return NewVerifyMFAUseCase(
		userRepo, sessionRepo, tokenRepo, tokenGen,
		encryptor, totpProvider, auditRepo,
		policies.SessionPolicy{ExpirationDays: 7},
	)
}

func TestVerifyMFAUseCase_Verify_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)
	tokenGen := mocks.NewTokenGenerator(t)
	encryptor := mocks.NewMFASecretEncryptor(t)
	totpProvider := mocks.NewTOTPProvider(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := buildVerifyMFAUseCase(
		t, userRepo, sessionRepo, tokenRepo, tokenGen, encryptor, totpProvider, auditRepo,
	)

	now := time.Now()
	userID := uuid.New().String()
	tokenID := uuid.New().String()
	rawToken := "rawtoken123"
	tokenHash := hashToken(rawToken)

	record := &entities.VerificationToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: domain.VerificationTokenTypeMFAChallenge,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
		UsedAt:    nil,
	}

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.UserWithMFA)
	user.ID = userID
	user.MFASecret = stringPtr("encrypted_secret")
	user.PasswordChangedAt = now
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Roles = []entities.Role{}

	authToken := authtest.CloneAuthToken(stubs.Entities.ValidAuthToken)
	authToken.AccessToken = "access_token"
	authToken.RefreshToken = "refresh_token"
	authToken.ExpiresAt = now.Add(time.Hour)

	tokenRepo.On("FindTokenByHash", ctx, tokenHash).Return(record, nil)
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	encryptor.On("Decrypt", ctx, "encrypted_secret").Return("plain_secret", nil)
	totpProvider.On("ValidateCode", ctx, "plain_secret", "123456").Return(true, nil)
	tokenRepo.On("MarkTokenAsUsed", ctx, tokenID).Return(nil)
	tokenGen.On("GenerateToken", user, mock.AnythingOfType("string")).Return(authToken, nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, userID, domain.SessionRevokeReasonNewLogin).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	// logLoginAttempt (called inside createSessionAndLoginResponse) looks up user by email
	userRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	req := &dtos.MFAVerifyRequest{ChallengeToken: rawToken, TOTPCode: "123456"}
	resp, err := uc.Verify(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Token)
	assert.Equal(t, authToken.AccessToken, resp.Token.AccessToken)
	assert.NotNil(t, resp.Session)
	assert.False(t, resp.MFARequired)

	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	encryptor.AssertExpectations(t)
	totpProvider.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestVerifyMFAUseCase_Verify_ChallengeTokenNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)
	tokenGen := mocks.NewTokenGenerator(t)
	encryptor := mocks.NewMFASecretEncryptor(t)
	totpProvider := mocks.NewTOTPProvider(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := buildVerifyMFAUseCase(
		t, userRepo, sessionRepo, tokenRepo, tokenGen, encryptor, totpProvider, auditRepo,
	)

	rawToken := "unknowntoken"
	tokenHash := hashToken(rawToken)
	tokenRepo.On("FindTokenByHash", ctx, tokenHash).Return((*entities.VerificationToken)(nil), errors.New("not found"))

	req := &dtos.MFAVerifyRequest{ChallengeToken: rawToken, TOTPCode: "000000"}
	resp, err := uc.Verify(ctx, req)

	assert.ErrorIs(t, err, domain.ErrMFAInvalidToken)
	assert.Nil(t, resp)
}

func TestVerifyMFAUseCase_Verify_TokenExpired(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)
	tokenGen := mocks.NewTokenGenerator(t)
	encryptor := mocks.NewMFASecretEncryptor(t)
	totpProvider := mocks.NewTOTPProvider(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := buildVerifyMFAUseCase(
		t, userRepo, sessionRepo, tokenRepo, tokenGen, encryptor, totpProvider, auditRepo,
	)

	rawToken := "expiredtoken"
	tokenHash := hashToken(rawToken)
	past := time.Now().Add(-time.Hour)
	record := &entities.VerificationToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		TokenHash: tokenHash,
		TokenType: domain.VerificationTokenTypeMFAChallenge,
		ExpiresAt: past,
		UsedAt:    nil,
	}
	tokenRepo.On("FindTokenByHash", ctx, tokenHash).Return(record, nil)

	req := &dtos.MFAVerifyRequest{ChallengeToken: rawToken, TOTPCode: "111111"}
	resp, err := uc.Verify(ctx, req)

	assert.ErrorIs(t, err, domain.ErrMFAInvalidToken)
	assert.Nil(t, resp)
}

func TestVerifyMFAUseCase_Verify_TokenAlreadyUsed(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)
	tokenGen := mocks.NewTokenGenerator(t)
	encryptor := mocks.NewMFASecretEncryptor(t)
	totpProvider := mocks.NewTOTPProvider(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := buildVerifyMFAUseCase(
		t, userRepo, sessionRepo, tokenRepo, tokenGen, encryptor, totpProvider, auditRepo,
	)

	rawToken := "usedtoken"
	tokenHash := hashToken(rawToken)
	usedAt := time.Now().Add(-time.Minute)
	record := &entities.VerificationToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		TokenHash: tokenHash,
		TokenType: domain.VerificationTokenTypeMFAChallenge,
		ExpiresAt: time.Now().Add(time.Hour),
		UsedAt:    &usedAt,
	}
	tokenRepo.On("FindTokenByHash", ctx, tokenHash).Return(record, nil)

	req := &dtos.MFAVerifyRequest{ChallengeToken: rawToken, TOTPCode: "222222"}
	resp, err := uc.Verify(ctx, req)

	assert.ErrorIs(t, err, domain.ErrMFAInvalidToken)
	assert.Nil(t, resp)
}

func TestVerifyMFAUseCase_Verify_InvalidTOTPCode(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)
	tokenGen := mocks.NewTokenGenerator(t)
	encryptor := mocks.NewMFASecretEncryptor(t)
	totpProvider := mocks.NewTOTPProvider(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := buildVerifyMFAUseCase(
		t, userRepo, sessionRepo, tokenRepo, tokenGen, encryptor, totpProvider, auditRepo,
	)

	now := time.Now()
	userID := uuid.New().String()
	rawToken := "validtoken"
	tokenHash := hashToken(rawToken)

	record := &entities.VerificationToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: domain.VerificationTokenTypeMFAChallenge,
		ExpiresAt: now.Add(time.Hour),
		UsedAt:    nil,
	}
	vstubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(vstubs.Entities.UserWithMFA)
	user.ID = userID
	user.MFASecret = stringPtr("encrypted_secret")
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Roles = []entities.Role{}

	tokenRepo.On("FindTokenByHash", ctx, tokenHash).Return(record, nil)
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	encryptor.On("Decrypt", ctx, "encrypted_secret").Return("plain_secret", nil)
	totpProvider.On("ValidateCode", ctx, "plain_secret", "999999").Return(false, nil)

	req := &dtos.MFAVerifyRequest{ChallengeToken: rawToken, TOTPCode: "999999"}
	resp, err := uc.Verify(ctx, req)

	assert.ErrorIs(t, err, domain.ErrMFAInvalidCode)
	assert.Nil(t, resp)
}

func TestVerifyMFAUseCase_Verify_UserNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)
	tokenGen := mocks.NewTokenGenerator(t)
	encryptor := mocks.NewMFASecretEncryptor(t)
	totpProvider := mocks.NewTOTPProvider(t)
	auditRepo := mocks.NewAuditRepository(t)

	uc := buildVerifyMFAUseCase(
		t, userRepo, sessionRepo, tokenRepo, tokenGen, encryptor, totpProvider, auditRepo,
	)

	now := time.Now()
	userID := uuid.New().String()
	rawToken := "tokenforemptyuser"
	tokenHash := hashToken(rawToken)

	record := &entities.VerificationToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: domain.VerificationTokenTypeMFAChallenge,
		ExpiresAt: now.Add(time.Hour),
		UsedAt:    nil,
	}

	tokenRepo.On("FindTokenByHash", ctx, tokenHash).Return(record, nil)
	userRepo.On("FindByID", ctx, userID).Return((*entities.User)(nil), errors.New("not found"))

	req := &dtos.MFAVerifyRequest{ChallengeToken: rawToken, TOTPCode: "333333"}
	resp, err := uc.Verify(ctx, req)

	assert.ErrorIs(t, err, domain.ErrMFAInvalidToken)
	assert.Nil(t, resp)
}
