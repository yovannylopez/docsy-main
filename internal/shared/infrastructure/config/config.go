package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/yovannylopez/docsy-main/pkg/config"
	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/databases"
)

// StorageConfig configures file storage (optional).
// Adjust DOCUMENT_PATH and MAX_FILE_SIZE in .env if the project requires file uploads.
type StorageConfig struct {
	DocumentPath string `json:"document_path"`
	MaxFileSize  int64  `json:"max_file_size"`
}

// RedisConfig configures distributed rate limiting (optional).
// If URL is empty, an in-memory limiter is used.
type RedisConfig struct {
	URL             string `json:"url"`
	AuthMaxRequests int    `json:"auth_max_requests"`
	AuthWindowSecs  int    `json:"auth_window_secs"`
}

// LDAPConfig configures authentication via LDAP / Active Directory (optional).
type LDAPConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	BaseDN  string `json:"base_dn"`
	BindDN  string `json:"bind_dn"`
	BindPW  string `json:"bind_password"`
}

// MFAConfig configures MFA TOTP settings.
type MFAConfig struct {
	// SecretKey is a 64-character hex-encoded 32-byte key for AES-GCM-256 encryption.
	// Must be provided via MFA_SECRET_KEY environment variable.
	SecretKey string `json:"secret_key"`
	// Issuer is the application name shown in authenticator apps.
	Issuer string `json:"issuer"`
}

// DBPoolConfig configures the connection pool and the database circuit breaker.
type DBPoolConfig struct {
	MaxOpenConns    int                            `json:"max_open_conns"`
	MaxIdleConns    int                            `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration                  `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration                  `json:"conn_max_idle_time"`
	ConnectTimeout  time.Duration                  `json:"connect_timeout"`
	QueryTimeout    time.Duration                  `json:"query_timeout"`
	MaxRetries      int                            `json:"max_retries"`
	RetryDelay      time.Duration                  `json:"retry_delay"`
	MaxBackoff      time.Duration                  `json:"max_backoff"`
	CircuitBreaker  databases.CircuitBreakerConfig `json:"circuit_breaker"`
}

// CoreConfig holds the complete application configuration.
// Extends pkg/config.BaseConfig with server-specific configuration.
type CoreConfig struct {
	config.BaseConfig
	Storage StorageConfig `json:"storage"`
	DBPool  DBPoolConfig  `json:"db_pool"`
	Redis   RedisConfig   `json:"redis"`
	LDAP    LDAPConfig    `json:"ldap"`
	MFA     MFAConfig     `json:"mfa"`
}

// NewCoreConfig loads configuration from the specified .env file and environment variables.
func NewCoreConfig(envFile string) (*CoreConfig, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("error loading env file: %w", err)
		}
	}

	baseConfig := config.GetBaseConfig()

	cfg := &CoreConfig{
		BaseConfig: baseConfig,
		Storage: StorageConfig{
			DocumentPath: config.GetEnv("DOCUMENT_PATH", "./storage/documents"),
			MaxFileSize: config.GetInt64Env(
				"MAX_FILE_SIZE",
				int64(constants.DefaultMaxFileSizeMB*constants.BytesPerMB),
			),
		},
		DBPool: getDBPoolConfig(),
		Redis: RedisConfig{
			URL:             config.GetEnv("REDIS_URL", ""),
			AuthMaxRequests: config.GetIntEnv("REDIS_AUTH_RATE_LIMIT", 60),       //nolint:mnd
			AuthWindowSecs:  config.GetIntEnv("REDIS_AUTH_RATE_WINDOW_SECS", 60), //nolint:mnd
		},
		LDAP: LDAPConfig{
			Enabled: config.GetBoolEnv("AUTH_LDAP_ENABLED", false),
			URL:     config.GetEnv("AUTH_LDAP_URL", ""),
			BaseDN:  config.GetEnv("AUTH_LDAP_BASE_DN", ""),
			BindDN:  config.GetEnv("AUTH_LDAP_BIND_DN", ""),
			BindPW:  config.GetEnv("AUTH_LDAP_BIND_PASSWORD", ""),
		},
		MFA: MFAConfig{
			SecretKey: config.GetEnv("MFA_SECRET_KEY", ""),
			Issuer:    config.GetEnv("MFA_ISSUER", "docsy-main"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return cfg, nil
}

// Validate checks that required fields are present and safe.
func (c *CoreConfig) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if strings.TrimSpace(c.Auth.JWTSecret) == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.Auth.JWTSecret == "your-secret-key" {
		return fmt.Errorf("JWT_SECRET must not use the insecure placeholder 'your-secret-key'")
	}

	if c.Auth.LockoutMaxAttempts < 0 {
		return fmt.Errorf("AUTH_LOCKOUT_MAX_ATTEMPTS must be >= 0")
	}
	if c.Auth.LockoutMaxAttempts > 0 && c.Auth.LockoutDuration <= 0 {
		return fmt.Errorf("AUTH_LOCKOUT_DURATION must be > 0 when AUTH_LOCKOUT_MAX_ATTEMPTS > 0")
	}

	return nil
}

// getDBPoolConfig returns the connection pool configuration based on the environment.
func getDBPoolConfig() DBPoolConfig {
	environment := config.GetEnv("ENVIRONMENT", constants.EnvDevelopment)

	var base databases.Config

	switch environment {
	case constants.EnvProduction:
		base = databases.ProductionConfig()
	case constants.EnvDevelopment:
		base = databases.DevelopmentConfig()
	default:
		base = databases.DefaultConfig()
	}

	return DBPoolConfig{
		MaxOpenConns:    base.MaxOpenConns,
		MaxIdleConns:    base.MaxIdleConns,
		ConnMaxLifetime: base.ConnMaxLifetime,
		ConnMaxIdleTime: base.ConnMaxIdleTime,
		ConnectTimeout:  base.ConnectTimeout,
		QueryTimeout:    base.QueryTimeout,
		MaxRetries:      base.MaxRetries,
		RetryDelay:      base.RetryDelay,
		MaxBackoff:      base.MaxBackoff,
		CircuitBreaker:  base.CircuitBreaker,
	}
}
