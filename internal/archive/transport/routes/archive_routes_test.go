package routes

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/transport/handlers"
	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

func TestRegisterArchiveRoutes(t *testing.T) {
	e := echo.New()
	g := e.Group("/api/v1")
	RegisterArchiveRoutes(g, &handlers.ArchiveHandler{})

	paths := make([]string, 0, len(e.Routes()))
	for _, route := range e.Routes() {
		paths = append(paths, route.Method+" "+route.Path)
	}
	require.Contains(t, paths, "GET /api/v1/archive/workspaces/me")
	require.Contains(t, paths, "GET /api/v1/archive/workspaces")
	require.Contains(t, paths, "POST /api/v1/archive/workspaces/household")
	require.Contains(t, paths, "GET /api/v1/archive/workspaces/:id/members")
	require.Contains(t, paths, "POST /api/v1/archive/workspaces/:id/members")
	require.Contains(t, paths, "PATCH /api/v1/archive/workspaces/:id/members/:userId")
	require.Contains(t, paths, "DELETE /api/v1/archive/workspaces/:id/members/:userId")
	require.Contains(t, paths, "GET /api/v1/archive/documents")
	require.Contains(t, paths, "POST /api/v1/archive/documents")
	require.Contains(t, paths, "GET /api/v1/archive/documents/:id")
	require.Contains(t, paths, "PATCH /api/v1/archive/documents/:id")
	require.Contains(t, paths, "POST /api/v1/archive/documents/:id/archive")
	require.Contains(t, paths, "GET /api/v1/archive/categories")
	require.Contains(t, paths, "POST /api/v1/archive/categories")
	require.Contains(t, paths, "PATCH /api/v1/archive/categories/:code")
	require.Contains(t, paths, "DELETE /api/v1/archive/categories/:code")
	require.Contains(t, paths, "GET /api/v1/archive/documents/:id/files")
	require.Contains(t, paths, "POST /api/v1/archive/documents/:id/files")
	require.Contains(t, paths, "GET /api/v1/archive/documents/:id/files/:fileId")
	require.Contains(t, paths, "DELETE /api/v1/archive/documents/:id/files/:fileId")
}

func TestWebArchiveRoutes_Setup(t *testing.T) {
	e := echo.New()
	wr := NewWebArchiveRoutes(&handlers.ArchivePageHandler{}, authmiddleware.NewWebAuthMiddleware(nil))
	wr.Setup(e)

	paths := make([]string, 0, len(e.Routes()))
	for _, route := range e.Routes() {
		paths = append(paths, route.Method+" "+route.Path)
	}
	require.Contains(t, paths, "GET /archivo")
	require.Contains(t, paths, "GET /archivo/hogares/nuevo")
	require.Contains(t, paths, "POST /archivo/hogares/nuevo")
	require.Contains(t, paths, "GET /archivo/hogares/:id/miembros")
	require.Contains(t, paths, "POST /archivo/hogares/:id/miembros")
	require.Contains(t, paths, "POST /archivo/hogares/:id/miembros/:userId/eliminar")
	require.Contains(t, paths, "GET /archivo/categorias")
	require.Contains(t, paths, "POST /archivo/categorias")
	require.Contains(t, paths, "POST /archivo/categorias/:code/editar")
	require.Contains(t, paths, "POST /archivo/categorias/:code/desactivar")
	require.Contains(t, paths, "GET /archivo/documentos")
	require.Contains(t, paths, "GET /archivo/documentos/nuevo")
	require.Contains(t, paths, "POST /archivo/documentos/nuevo")
	require.Contains(t, paths, "GET /archivo/documentos/:id/editar")
	require.Contains(t, paths, "POST /archivo/documentos/:id/editar")
	require.Contains(t, paths, "POST /archivo/documentos/:id/archivos")
	require.Contains(t, paths, "GET /archivo/documentos/:id/archivos/:fileId")
	require.Contains(t, paths, "POST /archivo/documentos/:id/archivos/:fileId/eliminar")
}
