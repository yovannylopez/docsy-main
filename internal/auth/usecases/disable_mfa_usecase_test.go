package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/mocks"
	authtest "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

func TestDisableMFAUseCase_Disable(t *testing.T) {
	const userID = "user-1"

	encSecret := "deadbeef"
	stubs := authtest.NewAuthStubs()
	mfaUser := authtest.CloneUser(stubs.Entities.UserWithMFA)
	mfaUser.ID = userID
	mfaUser.MFASecret = &encSecret

	tests := []struct {
		name      string
		req       *dtos.MFADisableRequest
		setupMock func(userRepo *mocks.UserRepository, totp *mocks.TOTPProvider, enc *mocks.MFASecretEncryptor)
		wantErr   error
	}{
		{
			name: "disables MFA when code is valid",
			req:  &dtos.MFADisableRequest{TOTPCode: "123456"},
			setupMock: func(userRepo *mocks.UserRepository, totp *mocks.TOTPProvider, enc *mocks.MFASecretEncryptor) {
				userRepo.On("FindByID", context.Background(), userID).Return(mfaUser, nil)
				enc.On("Decrypt", context.Background(), encSecret).Return("PLAIN_SECRET", nil)
				totp.On("ValidateCode", context.Background(), "PLAIN_SECRET", "123456").Return(true, nil)
				userRepo.On("DisableMFA", context.Background(), userID).Return(nil)
			},
		},
		{
			name: "returns ErrMFANotEnabled when MFA is inactive",
			req:  &dtos.MFADisableRequest{TOTPCode: "123456"},
			setupMock: func(userRepo *mocks.UserRepository, _ *mocks.TOTPProvider, _ *mocks.MFASecretEncryptor) {
				u := authtest.CloneUser(stubs.Entities.ValidUser)
				u.ID = userID
				u.MFAEnabled = false
				userRepo.On("FindByID", context.Background(), userID).Return(u, nil)
			},
			wantErr: domain.ErrMFANotEnabled,
		},
		{
			name: "returns ErrMFANotEnabled when user not found",
			req:  &dtos.MFADisableRequest{TOTPCode: "123456"},
			setupMock: func(userRepo *mocks.UserRepository, _ *mocks.TOTPProvider, _ *mocks.MFASecretEncryptor) {
				userRepo.On("FindByID", context.Background(), userID).Return(nil, errors.New("not found"))
			},
			wantErr: domain.ErrMFANotEnabled,
		},
		{
			name: "returns ErrMFAInvalidCode when TOTP code is wrong",
			req:  &dtos.MFADisableRequest{TOTPCode: "000000"},
			setupMock: func(userRepo *mocks.UserRepository, totp *mocks.TOTPProvider, enc *mocks.MFASecretEncryptor) {
				userRepo.On("FindByID", context.Background(), userID).Return(mfaUser, nil)
				enc.On("Decrypt", context.Background(), encSecret).Return("PLAIN_SECRET", nil)
				totp.On("ValidateCode", context.Background(), "PLAIN_SECRET", "000000").Return(false, nil)
			},
			wantErr: domain.ErrMFAInvalidCode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			totp := mocks.NewTOTPProvider(t)
			enc := mocks.NewMFASecretEncryptor(t)
			tc.setupMock(userRepo, totp, enc)

			uc := NewDisableMFAUseCase(userRepo, totp, enc)
			err := uc.Disable(context.Background(), userID, tc.req)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
