package routes

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

func TestWebAuditRoutes_Setup_RegistersRoutes(t *testing.T) {
	e := echo.New()
	wr := NewWebAuditRoutes(&handlers.AuditPageHandler{}, authmiddleware.NewWebAuthMiddleware(nil))
	wr.Setup(e)

	paths := make([]string, 0, len(e.Routes()))
	for _, route := range e.Routes() {
		paths = append(paths, route.Method+" "+route.Path)
	}
	require.Contains(t, paths, "GET /auditoria")
}
