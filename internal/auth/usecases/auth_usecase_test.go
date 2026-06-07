package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/infrastructure/security"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

func TestAuthUseCase_Authenticate_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	tokGen := mocks.NewTokenGenerator(t)

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.ValidUser)
	user.ID = "1"
	user.Email = "a@b.com"
	token := authtest.CloneAuthToken(stubs.Entities.ValidAuthToken)
	token.AccessToken = "a"
	token.RefreshToken = ""
	token.ExpiresAt = time.Now().Add(time.Hour)

	userRepo.On("FindByEmail", ctx, "a@b.com").Return(user, nil)
	tokGen.On("GenerateToken", user, "").Return(token, nil)

	uc := NewAuthUseCase(userRepo, tokGen, mocks.NewSessionRepository(t), nil)
	out, err := uc.Authenticate(ctx, "a@b.com", "ignored-password")

	require.NoError(t, err)
	assert.Equal(t, token, out)
}

func TestAuthUseCase_Authenticate_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	tokGen := mocks.NewTokenGenerator(t)

	userRepo.On("FindByEmail", ctx, "a@b.com").Return((*entities.User)(nil), errors.New("not found"))

	uc := NewAuthUseCase(userRepo, tokGen, mocks.NewSessionRepository(t), nil)
	out, err := uc.Authenticate(ctx, "a@b.com", "x")

	assert.Nil(t, out)
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestAuthUseCase_Authenticate_TokenError(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	tokGen := mocks.NewTokenGenerator(t)

	stubs := authtest.NewAuthStubs()
	user := authtest.CloneUser(stubs.Entities.ValidUser)
	user.ID = "1"
	user.Email = "a@b.com"
	userRepo.On("FindByEmail", ctx, "a@b.com").Return(user, nil)
	tokGen.On("GenerateToken", user, "").Return((*entities.AuthToken)(nil), errors.New("sign failed"))

	uc := NewAuthUseCase(userRepo, tokGen, mocks.NewSessionRepository(t), nil)
	out, err := uc.Authenticate(ctx, "a@b.com", "x")

	assert.Nil(t, out)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate token")
}

func TestAuthUseCase_ValidateToken_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessRepo := mocks.NewSessionRepository(t)
	tg := security.NewTokenGenerator("unit-test-secret-key-32bytes!!")
	stubs := authtest.NewAuthStubs()
	tokUser := authtest.CloneUser(stubs.Entities.ValidUser)
	tokUser.ID = "u1"
	tokUser.Email = "e@e.com"
	tokUser.Roles = []entities.Role{{Name: "funcionario"}}
	sessionID := "sess-1"
	at, err := tg.GenerateToken(tokUser, sessionID)
	require.NoError(t, err)

	dbUser := authtest.CloneUser(stubs.Entities.ValidUser)
	dbUser.ID = "u1"
	dbUser.Email = "e@e.com"
	dbUser.IsActive = true
	dbUser.PermissionNames = []string{"users.read"}
	userRepo.On("FindByID", ctx, "u1").Return(dbUser, nil)

	activeSession := authtest.CloneSession(stubs.Entities.ValidSession)
	activeSession.ID = sessionID
	activeSession.UserID = "u1"
	activeSession.IsActive = true
	activeSession.RevokedAt = nil
	activeSession.ExpiresAt = time.Now().Add(time.Hour)
	sessRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	uc := NewAuthUseCase(userRepo, tg, sessRepo, nil)
	u, err := uc.ValidateToken(ctx, at.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "u1", u.ID)
}

func TestAuthUseCase_ValidateToken_MissingSessionID(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessRepo := mocks.NewSessionRepository(t)
	tg := security.NewTokenGenerator("unit-test-secret-key-32bytes!!")
	stubs := authtest.NewAuthStubs()
	tokUser := authtest.CloneUser(stubs.Entities.ValidUser)
	tokUser.ID = "u1"
	at, err := tg.GenerateToken(tokUser, "")
	require.NoError(t, err)

	dbUser := authtest.CloneUser(stubs.Entities.ValidUser)
	dbUser.ID = "u1"
	dbUser.IsActive = true
	userRepo.On("FindByID", ctx, "u1").Return(dbUser, nil)

	uc := NewAuthUseCase(userRepo, tg, sessRepo, nil)
	_, err = uc.ValidateToken(ctx, at.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session_id missing")
}

func TestAuthUseCase_ValidateToken_RevokedSession(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewUserRepository(t)
	sessRepo := mocks.NewSessionRepository(t)
	tg := security.NewTokenGenerator("unit-test-secret-key-32bytes!!")
	stubs := authtest.NewAuthStubs()
	tokUser := authtest.CloneUser(stubs.Entities.ValidUser)
	tokUser.ID = "u1"
	sessionID := "sess-revoked"
	at, err := tg.GenerateToken(tokUser, sessionID)
	require.NoError(t, err)

	dbUser := authtest.CloneUser(stubs.Entities.ValidUser)
	dbUser.ID = "u1"
	dbUser.IsActive = true
	userRepo.On("FindByID", ctx, "u1").Return(dbUser, nil)

	revoked := authtest.CloneSession(stubs.Entities.RevokedSession)
	revoked.ID = sessionID
	revoked.UserID = "u1"
	sessRepo.On("FindByID", ctx, sessionID).Return(revoked, nil)

	uc := NewAuthUseCase(userRepo, tg, sessRepo, nil)
	_, err = uc.ValidateToken(ctx, at.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthUseCase_StubMethodsStillErrors(t *testing.T) {
	uc := NewAuthUseCase(mocks.NewUserRepository(t), mocks.NewTokenGenerator(t), mocks.NewSessionRepository(t), nil)
	ctx := context.Background()

	_, _, err := uc.Login(ctx, "e", "p", "ua", "ip")
	assert.Error(t, err)

	assert.Error(t, uc.LogoutAll(ctx, ""))
}
