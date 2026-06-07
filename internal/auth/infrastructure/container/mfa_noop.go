package container

import (
	"context"
	"errors"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
)

// noopMFASetup is a placeholder used when MFA_SECRET_KEY is not configured.
type noopMFASetup struct{}

func (noopMFASetup) Setup(_ context.Context, _ string) (*dtos.MFASetupResponse, error) {
	return nil, errors.New("MFA is not configured on this server")
}

// noopMFAConfirm is a placeholder used when MFA_SECRET_KEY is not configured.
type noopMFAConfirm struct{}

func (noopMFAConfirm) Confirm(_ context.Context, _ string, _ *dtos.MFAConfirmRequest) error {
	return errors.New("MFA is not configured on this server")
}

// noopMFAVerify is a placeholder used when MFA_SECRET_KEY is not configured.
type noopMFAVerify struct{}

func (noopMFAVerify) Verify(_ context.Context, _ *dtos.MFAVerifyRequest) (*dtos.LoginResponse, error) {
	return nil, errors.New("MFA is not configured on this server")
}

// noopMFADisable is a placeholder used when MFA_SECRET_KEY is not configured.
type noopMFADisable struct{}

func (noopMFADisable) Disable(_ context.Context, _ string, _ *dtos.MFADisableRequest) error {
	return errors.New("MFA is not configured on this server")
}
