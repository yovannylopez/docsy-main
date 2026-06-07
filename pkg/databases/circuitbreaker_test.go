package databases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appErrors "github.com/yovannylopez/docsy-main/pkg/errors"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

func init() {
	// Initialize the logger for the tests
	_ = logging.Init(false)
}

func TestNewCircuitBreakerWrapper(t *testing.T) {
	t.Run("creates wrapper with enabled circuit breaker", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            60 * time.Second,
			Timeout:             30 * time.Second,
			ConsecutiveFailures: 5,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)

		assert.NotNil(t, wrapper)
		assert.NotNil(t, wrapper.cb)
		assert.Equal(t, "closed", wrapper.GetState())
	})

	t.Run("creates wrapper with disabled circuit breaker", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			Enabled: false,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)

		assert.NotNil(t, wrapper)
		assert.Nil(t, wrapper.cb)
		assert.Equal(t, "disabled", wrapper.GetState())
	})
}

func TestCircuitBreakerWrapper_Execute(t *testing.T) {
	t.Run("executes successfully when circuit is closed", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            60 * time.Second,
			Timeout:             30 * time.Second,
			ConsecutiveFailures: 5,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)
		ctx := context.Background()

		executed := false
		err = wrapper.Execute(ctx, "test_operation", func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("propagates error from function", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            60 * time.Second,
			Timeout:             30 * time.Second,
			ConsecutiveFailures: 5,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)
		ctx := context.Background()

		expectedErr := errors.New("test error")
		err = wrapper.Execute(ctx, "test_operation", func() error {
			return expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("opens circuit after consecutive failures", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            1 * time.Second,
			Timeout:             2 * time.Second,
			ConsecutiveFailures: 3,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)
		ctx := context.Background()

		testErr := errors.New("database error")

		// Execute 3 times with error to open the circuit
		for i := 0; i < 3; i++ {
			err = wrapper.Execute(ctx, "test_operation", func() error {
				return testErr
			})
			assert.Error(t, err)
		}

		// Check if the circuit is open
		assert.Equal(t, "open", wrapper.GetState())

		// Try to execute with the circuit open
		err = wrapper.Execute(ctx, "test_operation", func() error {
			return nil
		})

		assert.Error(t, err)
		assert.True(t, appErrors.IsServiceUnavailableError(err))
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            60 * time.Second,
			Timeout:             30 * time.Second,
			ConsecutiveFailures: 5,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err = wrapper.Execute(ctx, "test_operation", func() error {
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	t.Run("executes directly when circuit breaker is disabled", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			Enabled: false,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)
		ctx := context.Background()

		executed := false
		err = wrapper.Execute(ctx, "test_operation", func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})
}

func TestCircuitBreakerWrapper_PingContext(t *testing.T) {
	t.Run("pings successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            60 * time.Second,
			Timeout:             30 * time.Second,
			ConsecutiveFailures: 5,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)
		ctx := context.Background()

		mock.ExpectPing()

		err = wrapper.PingContext(ctx)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates ping error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            60 * time.Second,
			Timeout:             30 * time.Second,
			ConsecutiveFailures: 5,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)
		ctx := context.Background()

		expectedErr := errors.New("ping failed")
		mock.ExpectPing().WillReturnError(expectedErr)

		err = wrapper.PingContext(ctx)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCircuitBreakerWrapper_GetCounts(t *testing.T) {
	t.Run("returns counts when enabled", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			MaxRequests:         3,
			Interval:            60 * time.Second,
			Timeout:             30 * time.Second,
			ConsecutiveFailures: 5,
			Enabled:             true,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)

		counts := wrapper.GetCounts()

		assert.NotNil(t, counts)
		assert.True(t, counts["enabled"].(bool))
		assert.Equal(t, "closed", counts["state"])
		assert.Equal(t, uint32(0), counts["requests"])
	})

	t.Run("returns disabled status when disabled", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		cfg := CircuitBreakerConfig{
			Enabled: false,
		}

		wrapper := NewCircuitBreakerWrapper(sqlxDB, cfg)

		counts := wrapper.GetCounts()

		assert.NotNil(t, counts)
		assert.False(t, counts["enabled"].(bool))
	})
}

func TestNewGobreakerCircuit(t *testing.T) {
	t.Run("disabled returns nil", func(t *testing.T) {
		cb := NewGobreakerCircuit("StandaloneTest", CircuitBreakerConfig{Enabled: false})
		assert.Nil(t, cb)
	})

	t.Run("enabled executes", func(t *testing.T) {
		cb := NewGobreakerCircuit("StandaloneTest", DefaultCircuitBreakerConfig())
		require.NotNil(t, cb)
		_, err := cb.Execute(func() (any, error) { return "ok", nil })
		assert.NoError(t, err)
	})
}

func TestCircuitBreakerConfig_Defaults(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := DefaultCircuitBreakerConfig()

		assert.Equal(t, uint32(3), cfg.MaxRequests)
		assert.Equal(t, 60*time.Second, cfg.Interval)
		assert.Equal(t, 30*time.Second, cfg.Timeout)
		assert.Equal(t, uint32(5), cfg.ConsecutiveFailures)
		assert.True(t, cfg.Enabled)
	})

	t.Run("production config", func(t *testing.T) {
		cfg := ProductionCircuitBreakerConfig()

		assert.Equal(t, uint32(5), cfg.MaxRequests)
		assert.Equal(t, 120*time.Second, cfg.Interval)
		assert.Equal(t, 60*time.Second, cfg.Timeout)
		assert.Equal(t, uint32(5), cfg.ConsecutiveFailures)
		assert.True(t, cfg.Enabled)
	})

	t.Run("development config", func(t *testing.T) {
		cfg := DevelopmentCircuitBreakerConfig()

		assert.Equal(t, uint32(2), cfg.MaxRequests)
		assert.Equal(t, 30*time.Second, cfg.Interval)
		assert.Equal(t, 15*time.Second, cfg.Timeout)
		assert.Equal(t, uint32(3), cfg.ConsecutiveFailures)
		assert.True(t, cfg.Enabled)
	})
}
