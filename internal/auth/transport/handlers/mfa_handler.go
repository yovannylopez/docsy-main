package handlers

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// MFAHandler handles all MFA TOTP HTTP operations.
type MFAHandler struct {
	setupService   ports.MFASetupService
	confirmService ports.MFAConfirmService
	verifyService  ports.MFAVerifyService
	disableService ports.MFADisableService
}

// NewMFAHandler creates a new MFAHandler.
func NewMFAHandler(
	setupService ports.MFASetupService,
	confirmService ports.MFAConfirmService,
	verifyService ports.MFAVerifyService,
	disableService ports.MFADisableService,
) *MFAHandler {
	return &MFAHandler{
		setupService:   setupService,
		confirmService: confirmService,
		verifyService:  verifyService,
		disableService: disableService,
	}
}

// SetupMFA handles POST /auth/mfa/setup (protected — requires authenticated user).
func (h *MFAHandler) SetupMFA(c echo.Context) error {
	user := userFromContext(c)
	if user == nil {
		return responses.Unauthorized(c, constants.LoginRequiredMessage)
	}

	result, err := h.setupService.Setup(c.Request().Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMFAAlreadyEnabled):
			return responses.BadRequest(c, constants.MFAAlreadyEnabledMessage)
		default:
			return responses.InternalError(c, constants.InternalErrorMessage)
		}
	}

	return responses.OK(c, result, constants.MFASetupInitiatedMessage)
}

// ConfirmMFA handles POST /auth/mfa/confirm (protected — requires authenticated user).
func (h *MFAHandler) ConfirmMFA(c echo.Context) error {
	user := userFromContext(c)
	if user == nil {
		return responses.Unauthorized(c, constants.LoginRequiredMessage)
	}

	var req dtos.MFAConfirmRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, constants.GenericErrorMessage)
	}
	if strings.TrimSpace(req.SetupToken) == "" || strings.TrimSpace(req.TOTPCode) == "" {
		return responses.BadRequest(c, constants.MFASetupTokenRequiredMessage)
	}

	if err := h.confirmService.Confirm(c.Request().Context(), user.ID, &req); err != nil {
		switch {
		case errors.Is(err, domain.ErrMFAInvalidToken):
			return responses.BadRequest(c, constants.GenericErrorMessage)
		case errors.Is(err, domain.ErrMFAInvalidCode):
			return responses.BadRequest(c, constants.GenericErrorMessage)
		default:
			return responses.InternalError(c, constants.InternalErrorMessage)
		}
	}

	return responses.OK(c, map[string]bool{"mfa_enabled": true}, constants.MFAEnabledMessage)
}

// VerifyMFA handles POST /auth/mfa/verify (public — called during the login flow).
func (h *MFAHandler) VerifyMFA(c echo.Context) error {
	var req dtos.MFAVerifyRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, constants.GenericErrorMessage)
	}
	if strings.TrimSpace(req.ChallengeToken) == "" || strings.TrimSpace(req.TOTPCode) == "" {
		return responses.BadRequest(c, constants.MFAChallengeTokenRequiredMessage)
	}

	result, err := h.verifyService.Verify(c.Request().Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMFAInvalidToken):
			return responses.BadRequest(c, constants.GenericErrorMessage)
		case errors.Is(err, domain.ErrMFAInvalidCode):
			return responses.BadRequest(c, constants.GenericErrorMessage)
		default:
			return responses.InternalError(c, constants.InternalErrorMessage)
		}
	}

	return responses.OK(c, result, constants.MFAVerifiedMessage)
}

// DisableMFA handles POST /auth/mfa/disable (protected — requires authenticated user).
func (h *MFAHandler) DisableMFA(c echo.Context) error {
	user := userFromContext(c)
	if user == nil {
		return responses.Unauthorized(c, constants.LoginRequiredMessage)
	}

	var req dtos.MFADisableRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, constants.GenericErrorMessage)
	}
	if strings.TrimSpace(req.TOTPCode) == "" {
		return responses.BadRequest(c, constants.MFATOTPCodeRequiredMessage)
	}

	if err := h.disableService.Disable(c.Request().Context(), user.ID, &req); err != nil {
		switch {
		case errors.Is(err, domain.ErrMFANotEnabled):
			return responses.BadRequest(c, constants.MFANotEnabledMessage)
		case errors.Is(err, domain.ErrMFAInvalidCode):
			return responses.BadRequest(c, constants.GenericErrorMessage)
		default:
			return responses.InternalError(c, constants.InternalErrorMessage)
		}
	}

	return responses.OK(c, map[string]bool{"mfa_disabled": true}, constants.MFADisabledMessage)
}
