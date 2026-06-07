package handlers

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/databases"
	"github.com/yovannylopez/docsy-main/pkg/errors"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db        *sqlx.DB
	cbWrapper *databases.CircuitBreakerWrapper
}

// NewHealthHandler creates a new instance of HealthHandler
func NewHealthHandler(db *sqlx.DB, cbWrapper *databases.CircuitBreakerWrapper) *HealthHandler {
	return &HealthHandler{
		db:        db,
		cbWrapper: cbWrapper,
	}
}

// HealthCheck checks the health status of the system
func (h *HealthHandler) HealthCheck(c echo.Context) error {
	// Check database health
	dbHealth := constants.HealthStatusHealthy
	dbError := ""

	if err := databases.HealthCheck(h.db, time.Duration(constants.HealthCheckTimeoutSeconds)*time.Second); err != nil {
		dbHealth = constants.HealthStatusUnhealthy
		dbError = err.Error()
		logging.Error("Database health check failed", logging.WithError(err))
	}

	// Get connection pool statistics
	dbStats := databases.GetConnectionStats(h.db)

	// Get Circuit Breaker statistics
	var cbStats map[string]any
	if h.cbWrapper != nil {
		cbStats = h.cbWrapper.GetCounts()
	}

	response := map[string]any{
		constants.JSONFieldStatus: "ok",
		"timestamp":               time.Now().UTC(),
		"services": map[string]any{
			"database": map[string]any{
				constants.JSONFieldStatus: dbHealth,
				constants.JSONFieldError:  dbError,
				"stats":                   dbStats,
				"circuit_breaker":         cbStats,
			},
			"api": map[string]any{
				constants.JSONFieldStatus: constants.HealthStatusHealthy,
			},
		},
	}

	statusCode := http_status.OK.Code
	if dbHealth == constants.HealthStatusUnhealthy {
		statusCode = http_status.ServiceUnavailable.Code
	}

	return errors.Wrap(c.JSON(statusCode, response), "failed to send health check response")
}

// ReadyCheck checks whether the system is ready to receive traffic
func (h *HealthHandler) ReadyCheck(c echo.Context) error {
	// Verify that the database is available
	if err := databases.HealthCheck(h.db, time.Duration(constants.ReadinessCheckTimeoutSeconds)*time.Second); err != nil {
		logging.Error("Readiness check failed", logging.WithError(err))

		return errors.Wrap(c.JSON(http_status.ServiceUnavailable.Code, map[string]any{
			constants.JSONFieldStatus: "not ready",
			constants.JSONFieldError:  "database not available",
			"message":                 constants.ServiceNotReadyToReceiveTrafficMessage,
		}), "failed to send readiness check error response")
	}

	return errors.Wrap(c.JSON(http_status.OK.Code, map[string]any{
		constants.JSONFieldStatus: "ready",
		"message":                 constants.ServiceReadyToReceiveTrafficMessage,
	}), "failed to send readiness check success response")
}
