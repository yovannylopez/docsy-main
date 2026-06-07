package routes

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
)

func TestAuthRoutes_Setup_RegistersAuthEndpoints(t *testing.T) {
	e := echo.New()
	ar := NewAuthRoutes(&handlers.AuthHandler{}, &handlers.MFAHandler{})
	g := e.Group("/api/v1/auth")
	ar.Setup(g)

	paths := routePaths(e)
	require.Contains(t, paths, "POST /api/v1/auth/login")
	require.NotContains(t, paths, "POST /api/v1/auth/signup")
	require.Contains(t, paths, "POST /api/v1/auth/refresh")
	require.Contains(t, paths, "POST /api/v1/auth/logout")
	require.Contains(t, paths, "POST /api/v1/auth/mfa/verify")
}

func TestAuthRoutes_SetupProtected_RegistersChangePasswordEndpoint(t *testing.T) {
	e := echo.New()
	ar := NewAuthRoutes(&handlers.AuthHandler{}, &handlers.MFAHandler{})
	g := e.Group("/api/v1/auth")
	ar.SetupProtected(g)

	paths := routePaths(e)
	require.Contains(t, paths, "POST /api/v1/auth/change-password")
	require.Contains(t, paths, "POST /api/v1/auth/mfa/setup")
	require.Contains(t, paths, "POST /api/v1/auth/mfa/confirm")
	require.Contains(t, paths, "POST /api/v1/auth/mfa/disable")
}

func TestAuditRoutes_Setup_RegistersAuditEndpoint(t *testing.T) {
	e := echo.New()
	ar := NewAuditRoutes(&handlers.AuditHandler{})
	g := e.Group("/api/v1")
	ar.Setup(g)

	paths := routePaths(e)
	require.Contains(t, paths, "GET /api/v1/auditoria")
}

func routePaths(e *echo.Echo) map[string]struct{} {
	out := make(map[string]struct{})
	for _, r := range e.Routes() {
		out[r.Method+" "+r.Path] = struct{}{}
	}
	return out
}

func TestNewAuthRoutes_NewAuditRoutes(t *testing.T) {
	ah := &handlers.AuthHandler{}
	mh := &handlers.MFAHandler{}
	authR := NewAuthRoutes(ah, mh)
	require.Equal(t, ah, authR.authHandler)
	require.Equal(t, mh, authR.mfaHandler)

	audH := &handlers.AuditHandler{}
	audR := NewAuditRoutes(audH)
	require.Equal(t, audH, audR.auditHandler)
}
