package databases

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

func init() {
	// Initialize the logger for the tests
	err := logging.Init(false) // false for development mode
	if err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 1*time.Minute, config.ConnMaxIdleTime)
	assert.Equal(t, 10*time.Second, config.ConnectTimeout)
	assert.Equal(t, 30*time.Second, config.QueryTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.MaxBackoff)
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	assert.Equal(t, 50, config.MaxOpenConns)
	assert.Equal(t, 25, config.MaxIdleConns)
	assert.Equal(t, 10*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, config.ConnMaxIdleTime)
	assert.Equal(t, 15*time.Second, config.ConnectTimeout)
	assert.Equal(t, 60*time.Second, config.QueryTimeout)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryDelay)
	assert.Equal(t, 60*time.Second, config.MaxBackoff)
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	assert.Equal(t, 10, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 2*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 30*time.Second, config.ConnMaxIdleTime)
	assert.Equal(t, 5*time.Second, config.ConnectTimeout)
	assert.Equal(t, 15*time.Second, config.QueryTimeout)
	assert.Equal(t, 2, config.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, config.RetryDelay)
	assert.Equal(t, 10*time.Second, config.MaxBackoff)
}

func TestNewConnection_InvalidConfig(t *testing.T) {
	config := Config{
		Host:     "invalid-host",
		Port:     "5432",
		User:     "test",
		Password: "test",
		DBName:   "test",
		SSLMode:  "disable",
		// Quick configuration to fail fast in tests
		MaxRetries:     1,
		RetryDelay:     10 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		ConnectTimeout: 100 * time.Millisecond,
	}

	ctx := context.Background()
	db, err := NewConnection(ctx, config)

	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "error connecting to database after 1 retries")
}

func TestGetConnectionStats(t *testing.T) {
	// Create a simulated connection for tests
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	stats := GetConnectionStats(db)

	// Verify all keys are present
	expectedKeys := []string{
		"max_open_connections",
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
		"wait_duration",
		"max_idle_closed",
		"max_lifetime_closed",
	}

	for _, key := range expectedKeys {
		assert.Contains(t, stats, key)
	}

	// Verify values are of the correct type
	assert.IsType(t, int(0), stats["max_open_connections"])
	assert.IsType(t, int(0), stats["open_connections"])
	assert.IsType(t, int(0), stats["in_use"])
	assert.IsType(t, int(0), stats["idle"])
	assert.IsType(t, int64(0), stats["wait_count"])
	assert.IsType(t, time.Duration(0), stats["wait_duration"])
	assert.IsType(t, int64(0), stats["max_idle_closed"])
	assert.IsType(t, int64(0), stats["max_lifetime_closed"])
}

func TestLogConnectionStats(t *testing.T) {
	// This function only logs, returns nothing
	// Only verify it doesn't cause panic
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Should not cause panic
	assert.NotPanics(t, func() {
		LogConnectionStats(db)
	})
}

func TestClose(t *testing.T) {
	// Create a simulated connection
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	// Close the connection
	err = Close(db)
	assert.NoError(t, err)

	// Try closing an already-closed connection
	// Some drivers may not return an error when closing an already-closed connection
	err = Close(db)
	// Verify it doesn't cause panic, but doesn't necessarily return an error
	assert.NotPanics(t, func() {
		_ = Close(db)
	})
}

func TestHealthCheck_Success(t *testing.T) {
	// Create a simulated connection
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Check health with short timeout
	err = HealthCheck(db, 1*time.Second)
	assert.NoError(t, err)
}

func TestHealthCheck_Failure(t *testing.T) {
	// Create a simulated connection and close it to simulate failure
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	db.Close()

	// Check health of a closed connection
	err = HealthCheck(db, 1*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database health check failed")
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "test",
				Password: "test",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "configuration without host",
			config: Config{
				Port:     "5432",
				User:     "test",
				Password: "test",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
		{
			name: "configuration without port",
			config: Config{
				Host:     "localhost",
				User:     "test",
				Password: "test",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate connection attempt
			ctx := context.Background()
			_, err := NewConnection(ctx, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// For valid configuration, we expect a connection error but not a configuration error
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error connecting to database")
			}
		})
	}
}

func TestConnectionRetry(t *testing.T) {
	config := Config{
		Host:           "invalid-host",
		Port:           "5432",
		User:           "test",
		Password:       "test",
		DBName:         "test",
		SSLMode:        "disable",
		MaxRetries:     2,
		RetryDelay:     10 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		ConnectTimeout: 50 * time.Millisecond,
	}

	ctx := context.Background()
	start := time.Now()
	_, err := NewConnection(ctx, config)
	duration := time.Since(start)

	// Verify the correct number of retries was attempted
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error connecting to database after 2 retries")

	// Verify total time includes retries with exponential backoff
	// With exponential backoff: attempt 1 = 10ms * 2^0 = 10ms, attempt 2 = 10ms * 2^1 = 20ms
	// With jitter (50-100%), the minimum would be: (10ms * 0.5) + (20ms * 0.5) = 5ms + 10ms = 15ms
	expectedMinDuration := 5 * time.Millisecond
	assert.GreaterOrEqual(t, duration, expectedMinDuration)
}

func TestCalculateBackoffWithJitter(t *testing.T) {
	tests := []struct {
		name       string
		attempt    int
		baseDelay  time.Duration
		maxBackoff time.Duration
		wantMin    time.Duration
		wantMax    time.Duration
	}{
		{
			name:       "attempt 0 returns 0",
			attempt:    0,
			baseDelay:  1 * time.Second,
			maxBackoff: 30 * time.Second,
			wantMin:    0,
			wantMax:    0,
		},
		{
			name:       "attempt 1 with jitter",
			attempt:    1,
			baseDelay:  1 * time.Second,
			maxBackoff: 30 * time.Second,
			wantMin:    500 * time.Millisecond, // 1s * 0.5
			wantMax:    1 * time.Second,        // 1s * 1.0
		},
		{
			name:       "attempt 2 with exponential growth",
			attempt:    2,
			baseDelay:  1 * time.Second,
			maxBackoff: 30 * time.Second,
			wantMin:    1 * time.Second, // 2s * 0.5
			wantMax:    2 * time.Second, // 2s * 1.0
		},
		{
			name:       "attempt 3 with exponential growth",
			attempt:    3,
			baseDelay:  1 * time.Second,
			maxBackoff: 30 * time.Second,
			wantMin:    2 * time.Second, // 4s * 0.5
			wantMax:    4 * time.Second, // 4s * 1.0
		},
		{
			name:       "respects max backoff",
			attempt:    10,
			baseDelay:  1 * time.Second,
			maxBackoff: 5 * time.Second,
			wantMin:    2500 * time.Millisecond, // 5s * 0.5
			wantMax:    5 * time.Second,         // capped at maxBackoff
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to verify the jitter range
			for i := 0; i < 10; i++ {
				got := calculateBackoffWithJitter(tt.attempt, tt.baseDelay, tt.maxBackoff)
				assert.GreaterOrEqual(t, got, tt.wantMin, "backoff should be >= minimum")
				assert.LessOrEqual(t, got, tt.wantMax, "backoff should be <= maximum")
			}
		})
	}
}

func TestNewConnection_ContextCancellation(t *testing.T) {
	config := Config{
		Host:           "invalid-host",
		Port:           "5432",
		User:           "test",
		Password:       "test",
		DBName:         "test",
		SSLMode:        "disable",
		MaxRetries:     5,
		RetryDelay:     100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		ConnectTimeout: 1 * time.Second,
	}

	// Create context that cancels after 200ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := NewConnection(ctx, config)
	duration := time.Since(start)

	// Verify it was cancelled by context
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection")

	// Verify it was cancelled before completing all retries
	// Without cancellation, it would take at least 5 attempts with exponential backoff
	maxExpectedDuration := 500 * time.Millisecond
	assert.Less(t, duration, maxExpectedDuration, "should cancel before all retries complete")
}

// Benchmark to measure function performance
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func BenchmarkProductionConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ProductionConfig()
	}
}

func BenchmarkDevelopmentConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DevelopmentConfig()
	}
}

func BenchmarkGetConnectionStats(b *testing.B) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetConnectionStats(db)
	}
}
