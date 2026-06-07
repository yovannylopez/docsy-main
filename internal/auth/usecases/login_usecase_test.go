package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	authports "github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

// runLoginUserReturnedButUseCaseFails covers Execute when FindByEmail returns a user but the use case rejects (locked, inactive, etc.).
func runLoginUserReturnedButUseCaseFails(t *testing.T, scen authtest.AuthLoginScenario, wantErr string) {
	t.Helper()
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)

	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	response, err := useCase.Execute(ctx, request, scen.UserAgent, scen.IPAddress)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, wantErr, err.Error())

	userRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.SuccessfulLogin
	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)
	token := authtest.CloneAuthToken(scen.Token)
	userAgent := scen.UserAgent
	ipAddress := scen.IPAddress

	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	// Setup expectations
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	tokenGenerator.On("GenerateToken", user, mock.AnythingOfType("string")).Return(token, nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, user.ID, domain.SessionRevokeReasonNewLogin).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Token)
	assert.NotNil(t, response.User)
	assert.NotNil(t, response.Session)

	// Verify token data
	assert.Equal(t, token.AccessToken, response.Token.AccessToken)
	assert.Equal(t, token.TokenType, response.Token.TokenType)
	assert.Equal(t, token.RefreshToken, response.Token.RefreshToken)

	// Verify user data
	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, user.Email, response.User.Email)
	assert.Equal(t, user.FirstName, response.User.FirstName)
	assert.Equal(t, user.LastName, response.User.LastName)
	assert.Equal(t, user.IsActive, response.User.IsActive)
	assert.Equal(t, user.IsVerified, response.User.IsVerified)
	assert.Len(t, response.User.Roles, 1)

	// Verify session data
	assert.Equal(t, user.ID, response.Session.UserID)
	assert.Equal(t, &userAgent, response.Session.UserAgent)
	assert.Equal(t, &ipAddress, response.Session.IPAddress)
	assert.True(t, response.Session.IsActive)

	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenGenerator.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	request := authtest.CloneLoginRequest(stubs.DTOs.ValidLoginRequest)
	request.Email = "nonexistent@example.com"
	userAgent := authtest.DefaultUserAgent
	ipAddress := authtest.DefaultIPAddress

	// Setup expectations
	userRepo.On("FindByEmail", ctx, request.Email).Return(nil, errors.New("user not found"))
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid credentials", err.Error())

	userRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_AccountLocked(t *testing.T) {
	stubs := authtest.NewAuthStubs()
	runLoginUserReturnedButUseCaseFails(t, stubs.Scenarios.AccountLocked, "account is locked")
}

func TestLoginUseCase_Execute_AccountNotActive(t *testing.T) {
	stubs := authtest.NewAuthStubs()
	runLoginUserReturnedButUseCaseFails(t, stubs.Scenarios.AccountInactive, "account is not active")
}

func TestLoginUseCase_Execute_InvalidPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.InvalidCredentials
	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)
	userAgent := scen.UserAgent
	ipAddress := scen.IPAddress

	// Setup expectations
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(false, nil)
	userRepo.On("RecordFailedPasswordAttempt", ctx, user.ID, 0, time.Duration(0)).
		Return(authports.FailedPasswordAttemptResult{FailedAttempts: 1, LockedUntil: nil}, nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid credentials", err.Error())

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_InvalidPasswordLockoutTriggered(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	lockout := policies.FailedLoginLockoutPolicy{MaxAttempts: 3, LockDuration: 15 * time.Minute}
	useCase := NewLoginUseCase(
		userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo,
		policies.SessionPolicy{ExpirationDays: 7}, nil, lockout,
	)

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.InvalidCredentials
	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)
	lu := time.Now().UTC().Add(lockout.LockDuration)

	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(false, nil)
	userRepo.On("RecordFailedPasswordAttempt", ctx, user.ID, 3, 15*time.Minute).
		Return(authports.FailedPasswordAttemptResult{FailedAttempts: 3, LockedUntil: &lu}, nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	response, err := useCase.Execute(ctx, request, scen.UserAgent, scen.IPAddress)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "account is locked", err.Error())

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
	_ = tokenGenerator
	_ = sessionRepo
}

func TestLoginUseCase_Execute_TokenGenerationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.SuccessfulLogin
	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)
	userAgent := scen.UserAgent
	ipAddress := scen.IPAddress

	// Setup expectations
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, user.ID, domain.SessionRevokeReasonNewLogin).Return(nil)
	tokenGenerator.On("GenerateToken", user, mock.AnythingOfType("string")).Return(nil, errors.New("token generation failed"))

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to generate authentication token")

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenGenerator.AssertExpectations(t)
}

func TestLoginUseCase_Execute_RevokePreviousSessionsError(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.SuccessfulLogin
	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)

	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, user.ID, domain.SessionRevokeReasonNewLogin).
		Return(errors.New("revoke failed"))

	response, err := useCase.Execute(ctx, request, scen.UserAgent, scen.IPAddress)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to revoke previous sessions")

	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
}

func TestLoginUseCase_Execute_SessionCreationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.SuccessfulLogin
	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)
	token := authtest.CloneAuthToken(scen.Token)
	userAgent := scen.UserAgent
	ipAddress := scen.IPAddress

	// Setup expectations
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	tokenGenerator.On("GenerateToken", user, mock.AnythingOfType("string")).Return(token, nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, user.ID, domain.SessionRevokeReasonNewLogin).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(errors.New("session creation failed"))

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create session")

	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenGenerator.AssertExpectations(t)
}

func TestLoginUseCase_buildLoginResponse(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.SuccessfulLogin
	token := authtest.CloneAuthToken(scen.Token)
	session := authtest.CloneSession(scen.Session)
	user := authtest.CloneUser(scen.User)

	// Act
	response := useCase.buildLoginResponse(token, session, user)

	// Assert
	assert.NotNil(t, response)
	assert.NotNil(t, response.Token)
	assert.NotNil(t, response.Session)
	assert.NotNil(t, response.User)

	// Verify token response
	assert.Equal(t, token.AccessToken, response.Token.AccessToken)
	assert.Equal(t, token.TokenType, response.Token.TokenType)
	assert.Equal(t, token.RefreshToken, response.Token.RefreshToken)

	// Verify session response
	assert.Equal(t, session.ID, response.Session.ID)
	assert.Equal(t, session.UserID, response.Session.UserID)
	assert.Equal(t, session.UserAgent, response.Session.UserAgent)
	assert.Equal(t, session.IPAddress, response.Session.IPAddress)
	assert.Equal(t, session.IsActive, response.Session.IsActive)

	// Verify user response
	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, user.Email, response.User.Email)
	assert.Equal(t, user.FirstName, response.User.FirstName)
	assert.Equal(t, user.LastName, response.User.LastName)
	assert.Equal(t, user.IsActive, response.User.IsActive)
	assert.Equal(t, user.IsVerified, response.User.IsVerified)
	assert.Len(t, response.User.Roles, 1)
	assert.Equal(t, "user", response.User.Roles[0].Name)
}

func TestLoginUseCase_Login(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.SuccessfulLogin
	user := authtest.CloneUser(scen.User)
	user.Roles = []entities.Role{}
	request := authtest.CloneLoginRequest(scen.Request)
	token := authtest.CloneAuthToken(scen.Token)
	userAgent := scen.UserAgent
	ipAddress := scen.IPAddress

	// Setup expectations
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	tokenGenerator.On("GenerateToken", user, mock.AnythingOfType("string")).Return(token, nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, user.ID, domain.SessionRevokeReasonNewLogin).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	// Act
	response, err := useCase.Login(ctx, request, userAgent, ipAddress)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)

	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenGenerator.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_logLoginAttempt(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.ValidUser)
	user.ID = "user123"
	email := user.Email
	ipAddress := authtest.DefaultIPAddress
	userAgent := authtest.DefaultUserAgent
	success := true
	sessionID := "session123"

	// Setup expectations
	userRepo.On("FindByEmail", ctx, email).Return(user, nil)
	auditRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionUserLoginAttempt &&
			*log.UserID == user.ID &&
			*log.SessionID == sessionID &&
			log.Result == domain.AuditResultSuccess &&
			*log.Message == email &&
			*log.IPAddress == ipAddress &&
			*log.UserAgent == userAgent
	})).Return(nil)

	// Act
	useCase.logLoginAttempt(ctx, email, ipAddress, userAgent, success, &sessionID)

	// Assert
	userRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_logLoginAttempt_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	email := "nonexistent@example.com"
	ipAddress := authtest.DefaultIPAddress
	userAgent := authtest.DefaultUserAgent
	success := false

	// Setup expectations - user not found
	userRepo.On("FindByEmail", ctx, email).Return(nil, errors.New("user not found"))
	auditRepo.On("LogAction", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.Action == domain.AuditActionUserLoginAttempt &&
			log.UserID == nil &&
			log.SessionID == nil &&
			log.Result == domain.AuditResultFailure &&
			*log.Message == email &&
			*log.IPAddress == ipAddress &&
			*log.UserAgent == userAgent
	})).Return(nil)

	// Act
	useCase.logLoginAttempt(ctx, email, ipAddress, userAgent, success, nil)

	// Assert
	userRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_PasswordVerificationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Scenarios.InvalidCredentials.User)
	request := authtest.CloneLoginRequest(stubs.Scenarios.InvalidCredentials.Request)
	request.Password = "password123"
	userAgent := authtest.DefaultUserAgent
	ipAddress := authtest.DefaultIPAddress

	// Setup expectations - password verification error
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(false, nil)
	userRepo.On("RecordFailedPasswordAttempt", ctx, user.ID, 0, time.Duration(0)).
		Return(authports.FailedPasswordAttemptResult{FailedAttempts: 1, LockedUntil: nil}, nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid credentials", err.Error())

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_SilentErrorHandling(t *testing.T) {
	// Test to verify silent error handling in non-critical operations
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.SuccessfulLogin
	user := authtest.CloneUser(scen.User)
	user.Roles = []entities.Role{}
	request := authtest.CloneLoginRequest(scen.Request)
	token := authtest.CloneAuthToken(scen.Token)
	userAgent := scen.UserAgent
	ipAddress := scen.IPAddress

	// Setup expectations with errors in non-critical operations
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(errors.New("reset failed"))
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(errors.New("update failed"))
	tokenGenerator.On("GenerateToken", user, mock.AnythingOfType("string")).Return(token, nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, user.ID, domain.SessionRevokeReasonNewLogin).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(errors.New("audit failed"))

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert - should continue despite silent errors
	assert.NoError(t, err)
	assert.NotNil(t, response)

	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenGenerator.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_IncrementFailedAttemptsError(t *testing.T) {
	// Test to verify silent error handling when incrementing failed attempts
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	scen := stubs.Scenarios.InvalidCredentials
	user := authtest.CloneUser(scen.User)
	request := authtest.CloneLoginRequest(scen.Request)
	userAgent := scen.UserAgent
	ipAddress := scen.IPAddress

	// Setup expectations
	userRepo.On("FindByEmail", ctx, request.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(false, nil)
	userRepo.On("RecordFailedPasswordAttempt", ctx, user.ID, 0, time.Duration(0)).
		Return(authports.FailedPasswordAttemptResult{}, errors.New("increment failed"))
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	// Act
	response, err := useCase.Execute(ctx, request, userAgent, ipAddress)

	// Assert - should fail due to invalid credentials, not the increment error
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid credentials", err.Error())

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLoginUseCase_buildLoginResponse_NilValues(t *testing.T) {
	// Test to verify nil value handling in buildLoginResponse
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenGenerator, auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{})

	stubs := authtest.NewAuthStubs()
	token := authtest.CloneAuthToken(stubs.Scenarios.SuccessfulLogin.Token)

	// Act with nil session and user
	response := useCase.buildLoginResponse(token, nil, nil)

	// Assert
	assert.NotNil(t, response)
	assert.NotNil(t, response.Token)
	assert.Nil(t, response.Session)
	assert.Nil(t, response.User)

	// Verify token response
	assert.Equal(t, token.AccessToken, response.Token.AccessToken)
	assert.Equal(t, token.TokenType, response.Token.TokenType)
	assert.Equal(t, token.RefreshToken, response.Token.RefreshToken)
}

// TestLoginUseCase_Execute_MFAEnabled verifies that when a user has MFA enabled,
// the login response contains MFARequired=true and a non-empty MFAChallengeToken
// instead of a full session (T025).
func TestLoginUseCase_Execute_MFAEnabled(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)

	useCase := NewLoginUseCase(
		userRepo, sessionRepo, passwordHasher, tokenGenerator,
		auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{},
	).WithMFA(tokenRepo, nil)

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.UserWithMFA)
	user.Roles = []entities.Role{}
	user.MFASecret = stringPtr("encrypted_secret")

	request := authtest.CloneLoginRequest(stubs.DTOs.ValidLoginRequest)
	request.Email = user.Email
	request.Password = "pass"

	userRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	tokenRepo.On(
		"CreateToken", ctx, mock.AnythingOfType("*entities.VerificationToken"),
	).Return(nil)

	response, err := useCase.Execute(ctx, request, authtest.DefaultUserAgent, authtest.DefaultIPAddress)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.MFARequired)
	assert.NotEmpty(t, response.MFAChallengeToken)
	assert.Nil(t, response.Token)
	assert.Nil(t, response.Session)
	assert.Nil(t, response.User)

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// TestLoginUseCase_Execute_MFAEnabled_TokenCreateError verifies error propagation
// when the token repository fails during MFA challenge creation (T025).
func TestLoginUseCase_Execute_MFAEnabled_TokenCreateError(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)
	tokenRepo := mocks.NewVerificationTokenRepository(t)

	useCase := NewLoginUseCase(
		userRepo, sessionRepo, passwordHasher, tokenGenerator,
		auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{},
	).WithMFA(tokenRepo, nil)

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.UserWithMFA)
	user.Roles = []entities.Role{}
	user.MFASecret = stringPtr("encrypted_secret")

	request := authtest.CloneLoginRequest(stubs.DTOs.ValidLoginRequest)
	request.Email = user.Email
	request.Password = "pass"

	userRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	tokenRepo.On(
		"CreateToken", ctx, mock.AnythingOfType("*entities.VerificationToken"),
	).Return(errors.New("db error"))

	response, err := useCase.Execute(ctx, request, authtest.DefaultUserAgent, authtest.DefaultIPAddress)

	assert.Error(t, err)
	assert.Nil(t, response)

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	_ = sessionRepo
	_ = tokenGenerator
	_ = auditRepo
}

// TestLoginUseCase_Execute_MFADisabled_NoTokenRepo is the regression check that
// when MFA is disabled (tokenRepo nil), the regular full-session flow still works (T025).
func TestLoginUseCase_Execute_MFADisabled_Regression(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessionRepo := mocks.NewSessionRepository(t)
	passwordHasher := mocks.NewPasswordHasher(t)
	tokenGenerator := mocks.NewTokenGenerator(t)
	auditRepo := mocks.NewAuditRepository(t)

	useCase := NewLoginUseCase(
		userRepo, sessionRepo, passwordHasher, tokenGenerator,
		auditRepo, policies.SessionPolicy{ExpirationDays: 7}, nil, policies.FailedLoginLockoutPolicy{},
	)

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.ValidUser)
	user.ID = "user-no-mfa"
	user.Email = "nomfa@example.com"
	user.Roles = []entities.Role{}
	token := authtest.CloneAuthToken(stubs.Scenarios.SuccessfulLogin.Token)
	token.AccessToken = "access_token"
	token.RefreshToken = "refresh_token"
	request := authtest.CloneLoginRequest(stubs.DTOs.ValidLoginRequest)
	request.Email = user.Email
	request.Password = "pass"

	userRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
	passwordHasher.On("VerifyPassword", request.Password, user.PasswordHash).Return(true, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, user.ID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	tokenGenerator.On("GenerateToken", user, mock.AnythingOfType("string")).Return(token, nil)
	sessionRepo.On("RevokeAllUserSessions", ctx, user.ID, domain.SessionRevokeReasonNewLogin).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	auditRepo.On("LogAction", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	response, err := useCase.Execute(ctx, request, authtest.DefaultUserAgent, authtest.DefaultIPAddress)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.False(t, response.MFARequired)
	assert.Empty(t, response.MFAChallengeToken)
	assert.NotNil(t, response.Token)
	assert.NotNil(t, response.Session)

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenGenerator.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}
