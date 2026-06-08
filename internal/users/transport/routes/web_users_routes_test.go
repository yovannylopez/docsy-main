package routes

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
	"github.com/yovannylopez/docsy-main/internal/users/transport/handlers"
)

func TestWebUsersRoutes_Setup_RegistersRoutes(t *testing.T) {
	e := echo.New()
	wr := NewWebUsersRoutes(&handlers.UsersPageHandler{}, authmiddleware.NewWebAuthMiddleware(nil))
	wr.Setup(e)

	paths := routePaths(e)
	require.Contains(t, paths, "GET /usuarios")
	require.Contains(t, paths, "GET /usuarios/nuevo")
	require.Contains(t, paths, "POST /usuarios")
	require.Contains(t, paths, "GET /usuarios/:id/editar")
	require.Contains(t, paths, "POST /usuarios/:id/editar")
}

func routePaths(e *echo.Echo) []string {
	paths := make([]string, 0, len(e.Routes()))
	for _, route := range e.Routes() {
		paths = append(paths, route.Method+" "+route.Path)
	}
	return paths
}
