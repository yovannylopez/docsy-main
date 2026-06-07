package routes

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

func TestWebAuthRoutes_Setup_RegistersPublicWebRoutes(t *testing.T) {
	e := echo.New()
	wr := NewWebAuthRoutes(&handlers.LoginPageHandler{}, authmiddleware.NewWebAuthMiddleware(nil), func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	})
	wr.Setup(e)

	paths := routePaths(e)
	require.Contains(t, paths, "GET /login")
	require.Contains(t, paths, "POST /login")
}

func TestWebAuthRoutes_SetupProtected_RegistersAuthenticatedWebRoutes(t *testing.T) {
	e := echo.New()
	wr := NewWebAuthRoutes(&handlers.LoginPageHandler{}, authmiddleware.NewWebAuthMiddleware(nil), nil)
	wr.SetupProtected(e)

	paths := routePaths(e)
	require.Contains(t, paths, "GET /")
	require.Contains(t, paths, "GET /perfil")
	require.Contains(t, paths, "GET /configuracion")
	require.Contains(t, paths, "POST /logout")
}

func TestNewWebAuthRoutes(t *testing.T) {
	h := &handlers.LoginPageHandler{}
	mw := authmiddleware.NewWebAuthMiddleware(nil)
	wr := NewWebAuthRoutes(h, mw, nil)
	require.Equal(t, h, wr.loginPageHandler)
	require.Equal(t, mw, wr.webAuthMW)
}
