// Package routes registers the HTTP routes for TEMPLATE_MODULE.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/infrastructure/container"
)

// TemplateRoutes holds the dependencies needed to register routes.
type TemplateRoutes struct {
	e         *echo.Echo
	container *container.TemplateContainer
}

// NewTemplateRoutes creates a new TemplateRoutes instance.
func NewTemplateRoutes(e *echo.Echo, c *container.TemplateContainer) *TemplateRoutes {
	return &TemplateRoutes{e: e, container: c}
}

// Register sets up all CRUD routes for the TEMPLATE_MODULE module.
// Pass JWT or other middleware as variadic arguments; they apply to all routes in the group.
func (r *TemplateRoutes) Register(middlewares ...echo.MiddlewareFunc) {
	g := r.e.Group("/api/TEMPLATE_MODULE", middlewares...)

	g.POST("", r.container.Handler.Create)
	g.GET("", r.container.Handler.List)
	g.GET("/:id", r.container.Handler.GetByID)
	g.PUT("/:id", r.container.Handler.Update)
	g.DELETE("/:id", r.container.Handler.Delete)
}
