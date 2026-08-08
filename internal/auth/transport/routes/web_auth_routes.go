package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

// WebAuthRoutes registers server-rendered HTML routes for authentication.
type WebAuthRoutes struct {
	loginPageHandler *handlers.LoginPageHandler
	webAuthMW        *authmiddleware.WebAuthMiddleware
	authRateLimit    echo.MiddlewareFunc
	extraMW          []echo.MiddlewareFunc
}

// NewWebAuthRoutes creates a WebAuthRoutes instance.
func NewWebAuthRoutes(
	loginPageHandler *handlers.LoginPageHandler,
	webAuthMW *authmiddleware.WebAuthMiddleware,
	authRateLimit echo.MiddlewareFunc,
	extraMW ...echo.MiddlewareFunc,
) *WebAuthRoutes {
	return &WebAuthRoutes{
		loginPageHandler: loginPageHandler,
		webAuthMW:        webAuthMW,
		authRateLimit:    authRateLimit,
		extraMW:          extraMW,
	}
}

// Setup registers public web auth routes (login).
func (wr *WebAuthRoutes) Setup(e *echo.Echo) {
	loginGroup := e.Group("")
	loginGroup.Use(wr.authRateLimit)
	loginGroup.GET("/login", wr.loginPageHandler.ShowLogin)
	loginGroup.POST("/login", wr.loginPageHandler.SubmitLogin)
}

// SetupProtected registers authenticated web routes (home, logout).
func (wr *WebAuthRoutes) SetupProtected(e *echo.Echo) {
	protected := e.Group("")
	protected.Use(wr.webAuthMW.RequireAuth())
	for _, mw := range wr.extraMW {
		protected.Use(mw)
	}
	protected.GET("/", wr.loginPageHandler.ShowHome)
	protected.GET("/perfil", wr.loginPageHandler.ShowProfile)
	protected.GET("/configuracion", wr.loginPageHandler.ShowSettings)
	protected.POST("/logout", wr.loginPageHandler.SubmitLogout)
}
