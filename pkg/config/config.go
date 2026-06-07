package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// AppConfig is the interface that must be implemented by the specific configurations
type AppConfig interface {
	Validate() error
}

// BaseConfig contains the common configuration for all services
type BaseConfig struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	Logging  LogConfig      `json:"logging"`
	Email    EmailConfig    `json:"email"`
}

// ServerConfig contains the server configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout"`
}

// DatabaseConfig contains the database configuration
type DatabaseConfig struct {
	// DATABASE_URL has priority over individual fields
	// Typical in PaaS and deployments that inject a single PostgreSQL URL
	DatabaseURL string `json:"databaseURL"`

	// Individual fields (used if DatabaseURL is empty)
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
	SSLMode  string `json:"sslMode"`
}

// AuthConfig contains the authentication configuration
type AuthConfig struct {
	JWTSecret       string        `json:"jwtSecret"`
	TokenExpiration time.Duration `json:"tokenExpiration"`
	RefreshDuration time.Duration `json:"refreshDuration"`
	// LockoutMaxAttempts: failed local password attempts before locked_until is set (0 = disable lockout only).
	LockoutMaxAttempts int `json:"lockoutMaxAttempts"`
	// LockoutDuration: how long locked_until stays in the future after threshold is reached.
	LockoutDuration time.Duration `json:"lockoutDuration"`
}

// LogConfig contains the logging configuration
type LogConfig struct {
	Level      string `json:"level"`
	OutputPath string `json:"outputPath"`
	Format     string `json:"format"`
}

// EmailConfig contains the email configuration for the email service
type EmailConfig struct {
	Provider   string `json:"provider"`
	APIKey     string `json:"apiKey"`
	FromEmail  string `json:"fromEmail"`
	FromName   string `json:"fromName"`
	DomainName string `json:"domainName"`
	Enabled    bool   `json:"enabled"`
}

// LoadConfig loads the configuration from the environment variables or the envFile (godotenv) and executes cfg.Validate().
// Does not assign fields of cfg from environment variables: build cfg with GetBaseConfig and the Get*Env helpers,
// or call godotenv.Load before reading the environment.
func LoadConfig[T AppConfig](envFile string, cfg T) (T, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return cfg, fmt.Errorf("error loading env file: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("config validation error: %w", err)
	}

	return cfg, nil
}

// GetEnv gets a value from the environment variables
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetIntEnv gets an integer value from the environment variables
func GetIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetInt64Env gets an integer value of 64 bits from the environment variables
func GetInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetBoolEnv gets a boolean value from the environment variables
func GetBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetDurationEnv gets a duration value from the environment variables
func GetDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetDatabaseConfig returns the database configuration
// Prioritizes DATABASE_URL (production / single URL) over individual variables (local)
func GetDatabaseConfig() DatabaseConfig {
	// Check first if DATABASE_URL exists (PaaS, containers, etc.)
	databaseURL := GetEnv("DATABASE_URL", "")

	if databaseURL != "" {
		// Single URL mode: use DATABASE_URL
		return DatabaseConfig{
			DatabaseURL: databaseURL,
			// Individual fields will be ignored, but we fill them in case
			Host:     GetEnv("DB_HOST", ""),
			Port:     GetEnv("DB_PORT", "5432"),
			User:     GetEnv("DB_USER", ""),
			Password: GetEnv("DB_PASSWORD", ""),
			DBName:   GetEnv("DB_NAME", ""),
			SSLMode:  GetEnv("DB_SSLMODE", "require"), // SSL required in production
		}
	}

	// Local mode: use individual variables
	return DatabaseConfig{
		DatabaseURL: "",
		Host:        GetEnv("DB_HOST", "localhost"),
		Port:        GetEnv("DB_PORT", "5432"),
		User:        GetEnv("DB_USER", "postgres"),
		Password:    GetEnv("DB_PASSWORD", ""),
		DBName:      GetEnv("DB_NAME", "go_boilerplate"),
		SSLMode:     GetEnv("DB_SSLMODE", "disable"), // SSL disabled in local
	}
}

// GetBaseConfig returns a common base configuration
func GetBaseConfig() BaseConfig {
	return BaseConfig{
		Server: ServerConfig{
			Port:         GetEnv("SERVER_PORT", "8100"),
			Host:         GetEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  GetDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: GetDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
		},
		Database: GetDatabaseConfig(),
		Auth: AuthConfig{
			// No default value: a fixed weak secret would be dangerous; CoreConfig.Validate requires explicit JWT_SECRET.
			JWTSecret:          GetEnv("JWT_SECRET", ""),
			TokenExpiration:    GetDurationEnv("TOKEN_EXPIRATION", 24*time.Hour),
			RefreshDuration:    GetDurationEnv("REFRESH_DURATION", 7*24*time.Hour),
			LockoutMaxAttempts: GetIntEnv("AUTH_LOCKOUT_MAX_ATTEMPTS", 3), //nolint:mnd
			LockoutDuration:    GetDurationEnv("AUTH_LOCKOUT_DURATION", 15*time.Minute),
		},
		Logging: LogConfig{
			Level:      GetEnv("LOG_LEVEL", "info"),
			OutputPath: GetEnv("LOG_OUTPUT", "stdout"),
			Format:     GetEnv("LOG_FORMAT", "json"),
		},
		Email: EmailConfig{
			Provider:   GetEnv("EMAIL_PROVIDER", "resend"),
			APIKey:     GetEnv("RESEND_API_KEY", ""),
			FromEmail:  GetEnv("EMAIL_FROM_ADDRESS", "noreply@example.com"),
			FromName:   GetEnv("EMAIL_FROM_NAME", "My App"),
			DomainName: GetEnv("EMAIL_DOMAIN", "example.com"),
			Enabled:    GetBoolEnv("EMAIL_ENABLED", false),
		},
	}
}
