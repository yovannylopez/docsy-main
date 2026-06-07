package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	authmw "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

// AuditRoutes handles the audit routes
type AuditRoutes struct {
	auditHandler *handlers.AuditHandler
}

// NewAuditRoutes creates a new AuditRoutes instance
func NewAuditRoutes(auditHandler *handlers.AuditHandler) *AuditRoutes {
	return &AuditRoutes{
		auditHandler: auditHandler,
	}
}

// Setup configures the audit routes (protected /api/v1 group).
func (r *AuditRoutes) Setup(g *echo.Group) {
	g.GET("/auditoria", r.auditHandler.List, authmw.RequirePermission("audit.read"))
}
