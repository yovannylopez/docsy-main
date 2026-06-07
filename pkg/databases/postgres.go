package databases

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"

	"github.com/yovannylopez/docsy-main/pkg/logging"
)

// Config contains the configuration needed to connect to PostgreSQL
type Config struct {
	// DATABASE_URL has priority over individual fields
	// Format: postgresql://user:password@host:port/database?sslmode=require
	DatabaseURL string `json:"database_url"`

	// Individual fields (used if DatabaseURL is empty)
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Connection pool configuration
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`

	// Timeout configuration
	ConnectTimeout time.Duration `json:"connect_timeout"`
	QueryTimeout   time.Duration `json:"query_timeout"`

	// Retry configuration
	MaxRetries int           `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`
	MaxBackoff time.Duration `json:"max_backoff"`

	// Circuit Breaker configuration
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`
}

// DefaultConfig returns a default optimized configuration
func DefaultConfig() Config {
	return Config{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
		ConnectTimeout:  10 * time.Second,
		QueryTimeout:    30 * time.Second,
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
		MaxBackoff:      30 * time.Second,
		CircuitBreaker:  DefaultCircuitBreakerConfig(),
	}
}

// ProductionConfig returns an optimized configuration for production
func ProductionConfig() Config {
	return Config{
		MaxOpenConns:    50,
		MaxIdleConns:    25,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnectTimeout:  15 * time.Second,
		QueryTimeout:    60 * time.Second,
		MaxRetries:      5,
		RetryDelay:      2 * time.Second,
		MaxBackoff:      60 * time.Second,
		CircuitBreaker:  ProductionCircuitBreakerConfig(),
	}
}

// DevelopmentConfig returns an optimized configuration for development
func DevelopmentConfig() Config {
	return Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 2 * time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
		ConnectTimeout:  5 * time.Second,
		QueryTimeout:    15 * time.Second,
		MaxRetries:      2,
		RetryDelay:      500 * time.Millisecond,
		MaxBackoff:      10 * time.Second,
		CircuitBreaker:  DevelopmentCircuitBreakerConfig(),
	}
}

// calculateBackoffWithJitter calculates the exponential backoff delay with jitter.
// Implements the Exponential Backoff with Jitter strategy to avoid the problem
// of 'thundering herd' during massive reconnections.
//
// Formula: min(base_delay * 2^(attempt-1), max_backoff) * (0.5 + random(0, 0.5))
//
// The jitter reduces the backoff between 50% and 100% of the calculated value, distributing
// the reconnections in time and reducing the load on the database.
func calculateBackoffWithJitter(attempt int, baseDelay, maxBackoff time.Duration) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Calculate exponential backoff: base_delay * 2^(attempt-1)
	backoff := baseDelay * time.Duration(1<<uint(attempt-1))

	// Limit to the maximum configured
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	// Add jitter: reduce between 50% and 100% of the value
	jitterFactor := 0.5 + rand.Float64()*0.5
	jitteredBackoff := time.Duration(float64(backoff) * jitterFactor)

	return jitteredBackoff
}

// ParseDatabaseURL parses a DATABASE_URL of PostgreSQL and extracts its components.
// Expected format: postgresql://user:password@host:port/database?sslmode=require
// Also supports the format: postgres://user:password@host:port/database
func ParseDatabaseURL(databaseURL string) (Config, error) {
	if databaseURL == "" {
		return Config{}, fmt.Errorf("database URL is empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Validate the scheme
	if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
		return Config{}, fmt.Errorf("invalid database URL scheme: %s (expected postgres or postgresql)", parsedURL.Scheme)
	}

	// Extract user and password
	user := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()

	// Extract host and port
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "5432" // Default port of PostgreSQL
	}

	// Extract database name
	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		return Config{}, fmt.Errorf("database name not found in URL")
	}

	// Extract query parameters (like sslmode)
	query := parsedURL.Query()
	sslMode := query.Get("sslmode")
	if sslMode == "" {
		sslMode = "require" // Default in URL without sslmode (typical in production)
	}

	logging.Info("Parsed DATABASE_URL successfully",
		zap.String("host", host),
		zap.String("port", port),
		zap.String("user", user),
		zap.String("database", dbName),
		zap.String("sslmode", sslMode))

	return Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}, nil
}

// NewConnection establishes a new connection with PostgreSQL with support for
// Exponential Backoff with Jitter in the retries.
//
// Supports two modes:
// 1. DATABASE_URL: If cfg.DatabaseURL is present, it is used as the main source
// 2. Individual fields: If DatabaseURL is empty, uses Host, Port, User, etc.
//
// The context allows canceling the retries of connection if the service stops.
// If ctx is nil, context.Background() is used internally.
func NewConnection(ctx context.Context, cfg Config) (*sqlx.DB, error) {
	// If no context is provided, use background
	if ctx == nil {
		ctx = context.Background()
	}
	// If DATABASE_URL is present, parse it and overwrite individual fields
	if cfg.DatabaseURL != "" {
		logging.Info("Using DATABASE_URL for connection")
		parsedConfig, err := ParseDatabaseURL(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
		}

		// Overwrite connection fields with the parsed ones
		cfg.Host = parsedConfig.Host
		cfg.Port = parsedConfig.Port
		cfg.User = parsedConfig.User
		cfg.Password = parsedConfig.Password
		cfg.DBName = parsedConfig.DBName
		cfg.SSLMode = parsedConfig.SSLMode
	}

	// Validate that we have the required fields
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("missing required database configuration fields")
	}

	// Log connection attempt
	logging.Info("Attempting database connection",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("user", cfg.User),
		zap.String("dbname", cfg.DBName),
		zap.String("sslmode", cfg.SSLMode),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns))

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Try connection with retry using Exponential Backoff with Jitter
	var db *sqlx.DB
	var err error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check if the context was cancelled before trying
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connection attempt cancelled before retry: %w", ctx.Err())
		default:
		}

		if attempt > 0 {
			// Calculate exponential backoff with jitter
			backoff := calculateBackoffWithJitter(attempt, cfg.RetryDelay, cfg.MaxBackoff)

			logging.Info("Retrying database connection with exponential backoff",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", cfg.MaxRetries),
				zap.Duration("backoff", backoff),
				zap.Duration("base_delay", cfg.RetryDelay),
				zap.Duration("max_backoff", cfg.MaxBackoff))

			// Wait with backoff, respecting the cancellation of the context
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("connection retry cancelled during backoff: %w", ctx.Err())
			case <-time.After(backoff):
				// Continue with the next attempt
			}
		}

		// Create context with timeout for this specific attempt
		connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
		defer cancel()

		db, err = sqlx.ConnectContext(connectCtx, "postgres", dsn)
		if err == nil {
			break
		}

		logging.Warn("Database connection attempt failed",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", cfg.MaxRetries))
	}

	if err != nil {
		logging.Error("Failed to connect to database after all retries",
			zap.Error(err),
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port),
			zap.String("user", cfg.User),
			zap.String("dbname", cfg.DBName),
			zap.String("sslmode", cfg.SSLMode))
		return nil, fmt.Errorf("error connecting to database after %d retries: %v", cfg.MaxRetries, err)
	}

	// Configure the connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Check the connection with ping
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	logging.Info("Database connection established successfully",
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cfg.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", cfg.ConnMaxIdleTime))

	return db, nil
}

// GetConnectionStats returns the statistics of the connection pool
func GetConnectionStats(db *sqlx.DB) map[string]any {
	if db == nil {
		return map[string]any{
			"max_open_connections": 0,
			"open_connections":     0,
			"in_use":               0,
			"idle":                 0,
			"wait_count":           0,
			"wait_duration":        0,
			"max_idle_closed":      0,
			"max_lifetime_closed":  0,
		}
	}

	stats := db.Stats()
	return map[string]any{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// LogConnectionStats logs the statistics of the connection pool
func LogConnectionStats(db *sqlx.DB) {
	// Check if the connection is not closed before getting statistics
	if db == nil {
		logging.Warn("Cannot log connection stats: database connection is nil")
		return
	}

	stats := GetConnectionStats(db)
	logging.Info("Database connection pool statistics",
		zap.Any("stats", stats))
}

// Close closes the connection with the database
func Close(db *sqlx.DB) error {
	// Log final statistics before closing
	LogConnectionStats(db)

	if err := db.Close(); err != nil {
		return fmt.Errorf("error closing database connection: %v", err)
	}
	logging.Info("Database connection closed successfully")
	return nil
}

// HealthCheck checks the health of the connection to the database
func HealthCheck(db *sqlx.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logging.Error("Database health check failed", zap.Error(err))
		return fmt.Errorf("database health check failed: %v", err)
	}

	logging.Debug("Database health check passed")
	return nil
}

// NewConnectionWithCircuitBreaker establishes a new connection with PostgreSQL
// protected by a Circuit Breaker.
// Returns both the database connection and the wrapper of the Circuit Breaker.
func NewConnectionWithCircuitBreaker(ctx context.Context, cfg Config) (*sqlx.DB, *CircuitBreakerWrapper, error) {
	// Establish the connection using NewConnection
	db, err := NewConnection(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	// Create the wrapper of the Circuit Breaker
	cbWrapper := NewCircuitBreakerWrapper(db, cfg.CircuitBreaker)

	// Check the connection with the Circuit Breaker
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := cbWrapper.PingContext(pingCtx); err != nil {
		// If the ping fails, close the connection
		_ = db.Close()
		return nil, nil, fmt.Errorf("failed to ping database through circuit breaker: %w", err)
	}

	logging.Info("Database connection with Circuit Breaker established successfully",
		zap.String("circuit_breaker_state", cbWrapper.GetState()),
		zap.Bool("circuit_breaker_enabled", cfg.CircuitBreaker.Enabled))

	return db, cbWrapper, nil
}
