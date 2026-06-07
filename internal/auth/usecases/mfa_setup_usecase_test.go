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
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

func TestSetupMFAUseCase_Setup(t *testing.T) {
	const userID = "user-1"

	type setupFn func(
		userRepo *mocks.UserRepository,
		tokenRepo *mocks.VerificationTokenRepository,
		totp *mocks.TOTPProvider,
		enc *mocks.MFASecretEncryptor,
	)

	stubs := authtest.NewAuthStubs()
	baseUser := func(mfaEnabled bool) *entities.User {
		u := authtest.CloneUser(stubs.Entities.ValidUser)
		u.ID = userID
		u.Email = "alice@example.com"
		u.IsActive = true
		u.MFAEnabled = mfaEnabled
		return u
	}

	tests := []struct {
		name      string
		setupMock setupFn
		wantErr   error
	}{
		{
			name: "returns setup response when MFA is not enabled",
			setupMock: func(
				userRepo *mocks.UserRepository,
				tokenRepo *mocks.VerificationTokenRepository,
				totp *mocks.TOTPProvider,
				enc *mocks.MFASecretEncryptor,
			) {
				userRepo.On("FindByID", context.Background(), userID).Return(baseUser(false), nil)
				totp.On("GenerateSecret", context.Background(), "", "alice@example.com").
					Return("SECRET", "otpauth://...", nil)
				enc.On("Encrypt", context.Background(), "SECRET").Return("encrypted", nil)
				userRepo.On("UpdateMFASecret", context.Background(), userID, "encrypted").Return(nil)
				tokenRepo.On("CreateToken", context.Background(), mock.Anything).Return(nil)
			},
		},
		{
			name: "returns ErrMFAAlreadyEnabled when MFA is active",
			setupMock: func(
				userRepo *mocks.UserRepository,
				_ *mocks.VerificationTokenRepository,
				_ *mocks.TOTPProvider,
				_ *mocks.MFASecretEncryptor,
			) {
				userRepo.On("FindByID", context.Background(), userID).Return(baseUser(true), nil)
			},
			wantErr: domain.ErrMFAAlreadyEnabled,
		},
		{
			name: "returns error when user not found",
			setupMock: func(
				userRepo *mocks.UserRepository,
				_ *mocks.VerificationTokenRepository,
				_ *mocks.TOTPProvider,
				_ *mocks.MFASecretEncryptor,
			) {
				userRepo.On("FindByID", context.Background(), userID).Return(nil, errors.New("not found"))
			},
			wantErr: domain.ErrMFAInvalidToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			tokenRepo := mocks.NewVerificationTokenRepository(t)
			totp := mocks.NewTOTPProvider(t)
			enc := mocks.NewMFASecretEncryptor(t)
			tc.setupMock(userRepo, tokenRepo, totp, enc)

			uc := NewSetupMFAUseCase(userRepo, tokenRepo, totp, enc)
			resp, err := uc.Setup(context.Background(), userID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotEmpty(t, resp.SetupToken)
			assert.Equal(t, "SECRET", resp.Secret)
		})
	}
}

func TestConfirmMFAUseCase_Confirm(t *testing.T) {
	const (
		userID   = "user-1"
		rawToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)

	encryptedSecret := "deadbeef"

	buildToken := func(used bool, expired bool) *entities.VerificationToken {
		exp := time.Now().Add(10 * time.Minute)
		if expired {
			exp = time.Now().Add(-1 * time.Minute)
		}
		var usedAt *time.Time
		if used {
			now := time.Now()
			usedAt = &now
		}
		return &entities.VerificationToken{
			ID:        "tok-1",
			UserID:    userID,
			TokenType: domain.VerificationTokenTypeMFASetup,
			ExpiresAt: exp,
			UsedAt:    usedAt,
			CreatedAt: time.Now(),
		}
	}

	confirmStubs := authtest.NewAuthStubs()
	baseUser := func() *entities.User {
		s := encryptedSecret
		u := authtest.CloneUser(confirmStubs.Entities.ValidUser)
		u.ID = userID
		u.MFAEnabled = false
		u.MFASecret = &s
		return u
	}

	tests := []struct {
		name      string
		req       *dtos.MFAConfirmRequest
		setupMock func(
			userRepo *mocks.UserRepository,
			tokenRepo *mocks.VerificationTokenRepository,
			totp *mocks.TOTPProvider,
			enc *mocks.MFASecretEncryptor,
		)
		wantErr error
	}{
		{
			name: "enables MFA when token and code are valid",
			req:  &dtos.MFAConfirmRequest{SetupToken: rawToken, TOTPCode: "123456"},
			setupMock: func(
				userRepo *mocks.UserRepository,
				tokenRepo *mocks.VerificationTokenRepository,
				totp *mocks.TOTPProvider,
				enc *mocks.MFASecretEncryptor,
			) {
				tokenRepo.On("FindTokenByHash", context.Background(), mock.Anything).
					Return(buildToken(false, false), nil)
				userRepo.On("FindByID", context.Background(), userID).Return(baseUser(), nil)
				enc.On("Decrypt", context.Background(), encryptedSecret).Return("PLAIN_SECRET", nil)
				totp.On("ValidateCode", context.Background(), "PLAIN_SECRET", "123456").Return(true, nil)
				tokenRepo.On("MarkTokenAsUsed", context.Background(), "tok-1").Return(nil)
				userRepo.On("EnableMFA", context.Background(), userID, encryptedSecret).Return(nil)
			},
		},
		{
			name: "returns ErrMFAInvalidToken when token not found",
			req:  &dtos.MFAConfirmRequest{SetupToken: rawToken, TOTPCode: "123456"},
			setupMock: func(
				_ *mocks.UserRepository,
				tokenRepo *mocks.VerificationTokenRepository,
				_ *mocks.TOTPProvider,
				_ *mocks.MFASecretEncryptor,
			) {
				tokenRepo.On("FindTokenByHash", context.Background(), mock.Anything).
					Return(nil, errors.New("not found"))
			},
			wantErr: domain.ErrMFAInvalidToken,
		},
		{
			name: "returns ErrMFAInvalidToken when token is already used",
			req:  &dtos.MFAConfirmRequest{SetupToken: rawToken, TOTPCode: "123456"},
			setupMock: func(
				_ *mocks.UserRepository,
				tokenRepo *mocks.VerificationTokenRepository,
				_ *mocks.TOTPProvider,
				_ *mocks.MFASecretEncryptor,
			) {
				tokenRepo.On("FindTokenByHash", context.Background(), mock.Anything).
					Return(buildToken(true, false), nil)
			},
			wantErr: domain.ErrMFAInvalidToken,
		},
		{
			name: "returns ErrMFAInvalidCode when TOTP code is wrong",
			req:  &dtos.MFAConfirmRequest{SetupToken: rawToken, TOTPCode: "000000"},
			setupMock: func(
				userRepo *mocks.UserRepository,
				tokenRepo *mocks.VerificationTokenRepository,
				totp *mocks.TOTPProvider,
				enc *mocks.MFASecretEncryptor,
			) {
				tokenRepo.On("FindTokenByHash", context.Background(), mock.Anything).
					Return(buildToken(false, false), nil)
				userRepo.On("FindByID", context.Background(), userID).Return(baseUser(), nil)
				enc.On("Decrypt", context.Background(), encryptedSecret).Return("PLAIN_SECRET", nil)
				totp.On("ValidateCode", context.Background(), "PLAIN_SECRET", "000000").Return(false, nil)
			},
			wantErr: domain.ErrMFAInvalidCode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			tokenRepo := mocks.NewVerificationTokenRepository(t)
			totp := mocks.NewTOTPProvider(t)
			enc := mocks.NewMFASecretEncryptor(t)
			tc.setupMock(userRepo, tokenRepo, totp, enc)

			uc := NewConfirmMFAUseCase(userRepo, tokenRepo, totp, enc)
			err := uc.Confirm(context.Background(), userID, tc.req)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
