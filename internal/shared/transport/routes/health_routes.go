package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/shared/transport/handlers"
)

// HealthRoutes es la estructura que contiene el handler de health checks
type HealthRoutes struct {
	healthHandler *handlers.HealthHandler
}

// NewHealthRoutes crea una nueva instancia de HealthRoutes
func NewHealthRoutes(healthHandler *handlers.HealthHandler) *HealthRoutes {
	return &HealthRoutes{
		healthHandler: healthHandler,
	}
}

// Setup configura las rutas de health checks
func (hr *HealthRoutes) Setup(g *echo.Group) {
	// Rutas de health checks
	g.GET("/health", hr.healthHandler.HealthCheck)
	g.GET("/ready", hr.healthHandler.ReadyCheck)
}
