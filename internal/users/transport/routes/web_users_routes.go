package routes

import (
	"github.com/labstack/echo/v4"

	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
	"github.com/yovannylopez/docsy-main/internal/users/transport/handlers"
)

// WebUsersRoutes registers server-rendered HTML routes for users.
type WebUsersRoutes struct {
	pageHandler *handlers.UsersPageHandler
	webAuthMW   *authmiddleware.WebAuthMiddleware
	extraMW     []echo.MiddlewareFunc
}

// NewWebUsersRoutes creates a WebUsersRoutes instance.
func NewWebUsersRoutes(
	pageHandler *handlers.UsersPageHandler,
	webAuthMW *authmiddleware.WebAuthMiddleware,
	extraMW ...echo.MiddlewareFunc,
) *WebUsersRoutes {
	return &WebUsersRoutes{
		pageHandler: pageHandler,
		webAuthMW:   webAuthMW,
		extraMW:     extraMW,
	}
}

// Setup registers authenticated users web routes.
func (wr *WebUsersRoutes) Setup(e *echo.Echo) {
	g := e.Group("")
	g.Use(wr.webAuthMW.RequireAuth())
	for _, mw := range wr.extraMW {
		g.Use(mw)
	}

	g.GET("/usuarios", wr.pageHandler.ListUsers, wr.webAuthMW.RequirePermission("users.read"))
	g.GET("/usuarios/nuevo", wr.pageHandler.ShowCreate, wr.webAuthMW.RequirePermission("users.create"))
	g.POST("/usuarios", wr.pageHandler.SubmitCreate, wr.webAuthMW.RequirePermission("users.create"))
	g.GET("/usuarios/:id/editar", wr.pageHandler.ShowEdit, wr.webAuthMW.RequirePermission("users.update"))
	g.POST("/usuarios/:id/editar", wr.pageHandler.SubmitEdit, wr.webAuthMW.RequirePermission("users.update"))
}
