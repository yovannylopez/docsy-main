package handlers

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

func bearerToken(c echo.Context) string {
	h := c.Request().Header.Get("Authorization")
	parts := strings.SplitN(h, " ", 2) //nolint:mnd
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// userFromContext extracts the authenticated user placed by AuthMiddleware.
func userFromContext(c echo.Context) *entities.User {
	u, _ := c.Get("user").(*entities.User)
	return u
}

// AuthHandler is the struct that holds the authentication services
type AuthHandler struct {
	loginService          ports.LoginService
	authService           ports.AuthenticationService
	changePasswordService ports.ChangePasswordService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(
	loginService ports.LoginService,
	authService ports.AuthenticationService,
	changePasswordService ports.ChangePasswordService,
) *AuthHandler {
	return &AuthHandler{
		loginService:          loginService,
		authService:           authService,
		changePasswordService: changePasswordService,
	}
}

// Login handles the login request
func (h *AuthHandler) Login(c echo.Context) error {
	var request dtos.LoginRequest

	// Parse the request JSON
	if err := c.Bind(&request); err != nil {
		return responses.BadRequest(c, constants.GenericErrorMessage)
	}

	request.Email = strings.TrimSpace(request.Email)

	if request.Email == "" || request.Password == "" {
		return responses.BadRequest(c, constants.LoginEmailPasswordRequiredMessage)
	}

	// Get client information
	userAgent := c.Request().UserAgent()
	ipAddress := h.getClientIP(c)

	// Execute the login use case
	responseData, err := h.loginService.Login(c.Request().Context(), &request, userAgent, ipAddress)
	if err != nil {
		// Handle specific errors
		switch err.Error() {
		case "invalid credentials":
			return responses.BadRequest(c, constants.InvalidCredentialsMessage)
		case "account is locked":
			return responses.BadRequest(c, "The account is temporarily locked")
		case "account is not active":
			return responses.BadRequest(c, "The account is not active")
		default:
			return responses.InternalError(c, constants.InternalErrorMessage)
		}
	}

	return responses.OK(c, responseData, constants.LoginSuccessMessage)
}

// RefreshToken maneja POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var body dtos.RefreshTokenRequest
	if err := c.Bind(&body); err != nil {
		return responses.BadRequest(c, constants.GenericErrorMessage)
	}
	if strings.TrimSpace(body.RefreshToken) == "" {
		return responses.BadRequest(c, "refresh_token is required")
	}

	tok, err := h.authService.RefreshToken(c.Request().Context(), body.RefreshToken)
	if err != nil {
		return responses.Unauthorized(c, "Invalid refresh token or revoked session")
	}

	return responses.OK(c, tok, "Token refreshed successfully")
}

// Logout maneja POST /api/v1/auth/logout (Bearer access o body refresh_token)
func (h *AuthHandler) Logout(c echo.Context) error {
	token := bearerToken(c)
	if token == "" {
		var body dtos.LogoutRequest
		if err := c.Bind(&body); err != nil {
			return responses.BadRequest(c, constants.GenericErrorMessage)
		}
		token = strings.TrimSpace(body.RefreshToken)
	}
	if token == "" {
		return responses.BadRequest(c, "Authorization Bearer or refresh_token in body is required")
	}

	if err := h.authService.Logout(
		c.Request().Context(),
		token,
		c.Request().UserAgent(),
		h.getClientIP(c),
	); err != nil {
		return responses.BadRequest(c, "Could not close the session")
	}

	return responses.OK(c, map[string]bool{"success": true}, "Session closed")
}

// ChangePassword handles POST /api/v1/auth/change-password.
// The user is identified exclusively via the Bearer JWT (FR-001).
// No tokens are returned in the response (FR-006).
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	user := userFromContext(c)
	if user == nil {
		return responses.Unauthorized(c, constants.LoginRequiredMessage)
	}

	var req dtos.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, constants.GenericErrorMessage)
	}

	if strings.TrimSpace(req.CurrentPassword) == "" || strings.TrimSpace(req.NewPassword) == "" {
		return responses.BadRequest(c, "current_password and new_password are required")
	}

	if err := h.changePasswordService.Execute(c.Request().Context(), user.ID, &req); err != nil {
		switch {
		case errors.Is(err, domain.ErrCurrentPasswordInvalid):
			// Generic message — must not reveal whether current password was wrong (FR-002 / T012).
			return responses.BadRequest(c, constants.ChangePasswordFailedMessage)
		case errors.Is(err, domain.ErrSamePassword):
			return responses.BadRequest(c, constants.ChangePasswordSameAsCurrentMsg)
		case errors.Is(err, domain.ErrPasswordInHistory):
			return responses.BadRequest(c, constants.PasswordInHistoryMessage)
		default:
			// Policy validation errors bubble up with their own message (US2 / T014).
			// Internal errors stay hidden behind the generic message.
			msg := err.Error()
			if msg == "password too short" ||
				msg == "password must contain uppercase letters" ||
				msg == "password must contain lowercase letters" ||
				msg == "password must contain numbers" ||
				msg == "password must contain symbols" {
				return responses.BadRequest(c, msg)
			}
			return responses.InternalError(c, constants.InternalErrorMessage)
		}
	}

	return responses.OK(c, dtos.ChangePasswordResponse{Message: constants.ChangePasswordSuccessMessage}, constants.ChangePasswordSuccessMessage)
}

// getClientIP gets the real IP address of the client
func (h *AuthHandler) getClientIP(c echo.Context) string {
	// Try to get the real IP considering proxies
	realIP := c.Request().Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	forwardedFor := c.Request().Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP from the list
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fallback to direct IP
	return c.RealIP()
}
