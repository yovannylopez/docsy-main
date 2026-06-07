package databases

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"

	"github.com/yovannylopez/docsy-main/pkg/errors"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

// CircuitBreakerConfig contains the configuration of the Circuit Breaker
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests uint32 `json:"max_requests"`

	// Interval is the period of time to reset the counter of failures in closed state
	Interval time.Duration `json:"interval"`

	// Timeout is the period of time that the circuit remains open before passing to half-open
	Timeout time.Duration `json:"timeout"`

	// ConsecutiveFailures is the number of consecutive failures before opening the circuit
	ConsecutiveFailures uint32 `json:"consecutive_failures"`

	// Enabled indicates if the Circuit Breaker is enabled
	Enabled bool `json:"enabled"`
}

// DefaultCircuitBreakerConfig returns the default configuration of the Circuit Breaker
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxRequests:         3,
		Interval:            60 * time.Second,
		Timeout:             30 * time.Second,
		ConsecutiveFailures: 5,
		Enabled:             true,
	}
}

// ProductionCircuitBreakerConfig returns the optimized configuration for production
func ProductionCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxRequests:         5,
		Interval:            120 * time.Second,
		Timeout:             60 * time.Second,
		ConsecutiveFailures: 5,
		Enabled:             true,
	}
}

// DevelopmentCircuitBreakerConfig returns the optimized configuration for development
func DevelopmentCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxRequests:         2,
		Interval:            30 * time.Second,
		Timeout:             15 * time.Second,
		ConsecutiveFailures: 3,
		Enabled:             true,
	}
}

// CircuitBreakerWrapper wraps a database connection with a Circuit Breaker
type CircuitBreakerWrapper struct {
	db *sqlx.DB
	cb *gobreaker.CircuitBreaker
}

// NewCircuitBreakerWrapper creates a new wrapper of Circuit Breaker for the database
func NewCircuitBreakerWrapper(db *sqlx.DB, cfg CircuitBreakerConfig) *CircuitBreakerWrapper {
	if !cfg.Enabled {
		logging.Info("Circuit Breaker is disabled, using direct database connection")
		return &CircuitBreakerWrapper{
			db: db,
			cb: nil,
		}
	}

	settings := gobreaker.Settings{
		Name:        "PostgreSQL",
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			shouldTrip := counts.ConsecutiveFailures >= cfg.ConsecutiveFailures

			if shouldTrip {
				logging.Error("Circuit Breaker is opening due to consecutive failures",
					zap.Uint32("consecutive_failures", counts.ConsecutiveFailures),
					zap.Uint32("threshold", cfg.ConsecutiveFailures),
					zap.Uint32("total_failures", counts.TotalFailures),
					zap.Uint32("total_requests", counts.Requests),
					zap.Float64("failure_ratio", failureRatio))
			}

			return shouldTrip
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logging.Info("Circuit Breaker state changed",
				zap.String("circuit", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}

	cb := gobreaker.NewCircuitBreaker(settings)

	logging.Info("Circuit Breaker initialized",
		zap.Uint32("max_requests", cfg.MaxRequests),
		zap.Duration("interval", cfg.Interval),
		zap.Duration("timeout", cfg.Timeout),
		zap.Uint32("consecutive_failures_threshold", cfg.ConsecutiveFailures))

	return &CircuitBreakerWrapper{
		db: db,
		cb: cb,
	}
}

// NewGobreakerCircuit creates a standalone circuit breaker (e.g. HTTP client or workers).
// Returns nil if cfg.Enabled is false.
func NewGobreakerCircuit(name string, cfg CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	if !cfg.Enabled {
		return nil
	}

	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.ConsecutiveFailures
		},
		OnStateChange: func(n string, from gobreaker.State, to gobreaker.State) {
			logging.Info("Circuit Breaker state changed",
				zap.String("circuit", n),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}

	return gobreaker.NewCircuitBreaker(settings)
}

// DB returns the underlying database instance
func (cbw *CircuitBreakerWrapper) DB() *sqlx.DB {
	return cbw.db
}

// Execute executes a function with protection of the Circuit Breaker
func (cbw *CircuitBreakerWrapper) Execute(ctx context.Context, operation string, fn func() error) error {
	// If the Circuit Breaker is disabled, execute directly
	if cbw.cb == nil {
		return fn()
	}

	// Execute with Circuit Breaker
	_, err := cbw.cb.Execute(func() (any, error) {
		// Check if the context was cancelled
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Execute the function
		if err := fn(); err != nil {
			logging.Error("Database operation failed",
				zap.String("operation", operation),
				zap.Error(err))
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		// If the error is because the circuit is open, return typed error
		if err == gobreaker.ErrOpenState {
			logging.Warn("Circuit Breaker is open, rejecting database operation",
				zap.String("operation", operation),
				zap.String("state", cbw.cb.State().String()))

			return errors.ServiceUnavailableError(
				"Database",
				"Circuit breaker is open due to repeated failures",
			)
		}

		// If the error is because there are too many requests in half-open, return typed error
		if err == gobreaker.ErrTooManyRequests {
			logging.Warn("Circuit Breaker rejected request: too many requests in half-open state",
				zap.String("operation", operation),
				zap.String("state", cbw.cb.State().String()))

			return errors.ServiceUnavailableError(
				"Database",
				"Too many requests while circuit breaker is recovering",
			)
		}

		// For other errors, return them as is
		return err
	}

	return nil
}

// PingContext checks the connection with protection of the Circuit Breaker
func (cbw *CircuitBreakerWrapper) PingContext(ctx context.Context) error {
	return cbw.Execute(ctx, "ping", func() error {
		return cbw.db.PingContext(ctx)
	})
}

// GetState returns the current state of the Circuit Breaker
func (cbw *CircuitBreakerWrapper) GetState() string {
	if cbw.cb == nil {
		return "disabled"
	}
	return cbw.cb.State().String()
}

// GetCounts returns the statistics of the Circuit Breaker
func (cbw *CircuitBreakerWrapper) GetCounts() map[string]any {
	if cbw.cb == nil {
		return map[string]any{
			"enabled": false,
		}
	}

	counts := cbw.cb.Counts()
	return map[string]any{
		"enabled":               true,
		"state":                 cbw.cb.State().String(),
		"requests":              counts.Requests,
		"total_successes":       counts.TotalSuccesses,
		"total_failures":        counts.TotalFailures,
		"consecutive_successes": counts.ConsecutiveSuccesses,
		"consecutive_failures":  counts.ConsecutiveFailures,
	}
}

// LogCircuitBreakerStats logs the statistics of the Circuit Breaker
func (cbw *CircuitBreakerWrapper) LogCircuitBreakerStats() {
	if cbw.cb == nil {
		logging.Debug("Circuit Breaker is disabled")
		return
	}

	stats := cbw.GetCounts()
	logging.Info("Circuit Breaker statistics", zap.Any("stats", stats))
}
