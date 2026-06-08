package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

// WebAuditRoutes registers server-rendered HTML routes for audit logs.
type WebAuditRoutes struct {
	pageHandler *handlers.AuditPageHandler
	webAuthMW   *authmiddleware.WebAuthMiddleware
}

// NewWebAuditRoutes creates a WebAuditRoutes instance.
func NewWebAuditRoutes(pageHandler *handlers.AuditPageHandler, webAuthMW *authmiddleware.WebAuthMiddleware) *WebAuditRoutes {
	return &WebAuditRoutes{
		pageHandler: pageHandler,
		webAuthMW:   webAuthMW,
	}
}

// Setup registers authenticated audit web routes.
func (wr *WebAuditRoutes) Setup(e *echo.Echo) {
	g := e.Group("")
	g.Use(wr.webAuthMW.RequireAuth())
	g.GET("/auditoria", wr.pageHandler.List, wr.webAuthMW.RequirePermission("audit.read"))
}
