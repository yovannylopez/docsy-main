package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
)

// AuthRoutes is the struct that holds the authentication handler
type AuthRoutes struct {
	authHandler *handlers.AuthHandler
	mfaHandler  *handlers.MFAHandler
}

// NewAuthRoutes creates a new AuthRoutes instance
func NewAuthRoutes(authHandler *handlers.AuthHandler, mfaHandler *handlers.MFAHandler) *AuthRoutes {
	return &AuthRoutes{
		authHandler: authHandler,
		mfaHandler:  mfaHandler,
	}
}

// Setup configures the public (unauthenticated) authentication routes
func (ar *AuthRoutes) Setup(g *echo.Group) {
	// Authentication routes
	g.POST("/login", ar.authHandler.Login)
	g.POST("/refresh", ar.authHandler.RefreshToken)
	g.POST("/logout", ar.authHandler.Logout)

	// MFA public route (called during the two-step login flow)
	g.POST("/mfa/verify", ar.mfaHandler.VerifyMFA)
}

// SetupProtected registers routes that require an authenticated user.
func (ar *AuthRoutes) SetupProtected(g *echo.Group) {
	g.POST("/change-password", ar.authHandler.ChangePassword)

	// MFA protected routes
	g.POST("/mfa/setup", ar.mfaHandler.SetupMFA)
	g.POST("/mfa/confirm", ar.mfaHandler.ConfirmMFA)
	g.POST("/mfa/disable", ar.mfaHandler.DisableMFA)
}
