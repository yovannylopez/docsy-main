package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/pkg/config"
	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/databases"
)

func TestNewCoreConfig(t *testing.T) {
	tests := []struct {
		name        string
		envFile     string
		envVars     map[string]string
		expectError bool
		checkConfig func(*testing.T, *CoreConfig)
	}{
		{
			name:    "successful config creation with valid env file",
			envFile: ".env.test",
			envVars: map[string]string{
				"ENVIRONMENT":   constants.EnvDevelopment,
				"DOCUMENT_PATH": "/test/documents",
				"MAX_FILE_SIZE": "10485760", // 10MB in bytes
			},
			expectError: false,
			checkConfig: func(t *testing.T, cfg *CoreConfig) {
				assert.Equal(t, "/test/documents", cfg.Storage.DocumentPath)
				assert.Equal(t, int64(10485760), cfg.Storage.MaxFileSize)
				assert.NotNil(t, cfg.DBPool)
			},
		},
		{
			name:        "error when env file does not exist",
			envFile:     "nonexistent.env",
			expectError: true,
		},
		{
			name:    "successful config with default values",
			envFile: ".env.test",
			envVars: map[string]string{
				"ENVIRONMENT": constants.EnvDevelopment,
			},
			expectError: false,
			checkConfig: func(t *testing.T, cfg *CoreConfig) {
				assert.Equal(t, "./storage/documents", cfg.Storage.DocumentPath)
				assert.Equal(t, int64(constants.DefaultMaxFileSizeMB*constants.BytesPerMB), cfg.Storage.MaxFileSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			if tt.envVars != nil {
				for key, value := range tt.envVars {
					_ = os.Setenv(key, value)
				}

				defer func() {
					for key := range tt.envVars {
						_ = os.Unsetenv(key)
					}
				}()
			}

			// Create test .env file if needed
			if tt.envFile == ".env.test" {
				createTestEnvFile(t, tt.envFile)

				defer func() {
					_ = os.Remove(tt.envFile)
				}()
			}

			// Execute test
			cfg, err := NewCoreConfig(tt.envFile)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)

				if tt.checkConfig != nil {
					tt.checkConfig(t, cfg)
				}
			}
		})
	}
}

func TestGetDBPoolConfig(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    databases.Config
	}{
		{
			name:        "production environment",
			environment: "production",
			expected:    databases.ProductionConfig(),
		},
		{
			name:        "development environment",
			environment: constants.EnvDevelopment,
			expected:    databases.DevelopmentConfig(),
		},
		{
			name:        "default environment",
			environment: "unknown",
			expected:    databases.DefaultConfig(),
		},
		{
			name:        "empty environment",
			environment: "",
			expected:    databases.DevelopmentConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			_ = os.Setenv("ENVIRONMENT", tt.environment)
			defer func() {
				_ = os.Unsetenv("ENVIRONMENT")
			}()

			// Execute function
			result := getDBPoolConfig()

			// Assertions
			assert.Equal(t, tt.expected.MaxOpenConns, result.MaxOpenConns)
			assert.Equal(t, tt.expected.MaxIdleConns, result.MaxIdleConns)
			assert.Equal(t, tt.expected.ConnMaxLifetime, result.ConnMaxLifetime)
			assert.Equal(t, tt.expected.ConnMaxIdleTime, result.ConnMaxIdleTime)
			assert.Equal(t, tt.expected.ConnectTimeout, result.ConnectTimeout)
			assert.Equal(t, tt.expected.QueryTimeout, result.QueryTimeout)
			assert.Equal(t, tt.expected.MaxRetries, result.MaxRetries)
			assert.Equal(t, tt.expected.RetryDelay, result.RetryDelay)
		})
	}
}

func TestCoreConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *CoreConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: &CoreConfig{
				BaseConfig: config.BaseConfig{
					Server: config.ServerConfig{Port: "8080"},
					Auth:   config.AuthConfig{JWTSecret: "test-secret"},
				},
			},
			expectError: false,
		},
		{
			name: "missing server port",
			config: &CoreConfig{
				BaseConfig: config.BaseConfig{
					Server: config.ServerConfig{Port: ""},
					Auth:   config.AuthConfig{JWTSecret: "test-secret"},
				},
			},
			expectError: true,
			errorMsg:    "server port is required",
		},
		{
			name: "missing JWT secret",
			config: &CoreConfig{
				BaseConfig: config.BaseConfig{
					Server: config.ServerConfig{Port: "8080"},
					Auth:   config.AuthConfig{JWTSecret: ""},
				},
			},
			expectError: true,
			errorMsg:    "JWT secret is required",
		},
		{
			name: "insecure legacy JWT placeholder",
			config: &CoreConfig{
				BaseConfig: config.BaseConfig{
					Server: config.ServerConfig{Port: "8080"},
					Auth:   config.AuthConfig{JWTSecret: "your-secret-key"},
				},
			},
			expectError: true,
			errorMsg:    "JWT_SECRET must not use the insecure placeholder",
		},
		{
			name: "multiple missing required fields",
			config: &CoreConfig{
				BaseConfig: config.BaseConfig{
					Server: config.ServerConfig{Port: ""},
					Auth:   config.AuthConfig{JWTSecret: ""},
				},
			},
			expectError: true,
			errorMsg:    "server port is required",
		},
		{
			name: "negative AUTH_LOCKOUT_MAX_ATTEMPTS",
			config: &CoreConfig{
				BaseConfig: config.BaseConfig{
					Server: config.ServerConfig{Port: "8080"},
					Auth:   config.AuthConfig{JWTSecret: "test-secret", LockoutMaxAttempts: -1},
				},
			},
			expectError: true,
			errorMsg:    "AUTH_LOCKOUT_MAX_ATTEMPTS must be >= 0",
		},
		{
			name: "lockout max set without positive duration",
			config: &CoreConfig{
				BaseConfig: config.BaseConfig{
					Server: config.ServerConfig{Port: "8080"},
					Auth: config.AuthConfig{
						JWTSecret:          "test-secret",
						LockoutMaxAttempts: 3,
						LockoutDuration:    0,
					},
				},
			},
			expectError: true,
			errorMsg:    "AUTH_LOCKOUT_DURATION must be > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStorageConfig(t *testing.T) {
	t.Run("storage config creation", func(t *testing.T) {
		storage := StorageConfig{
			DocumentPath: "/test/path",
			MaxFileSize:  10485760,
		}

		assert.Equal(t, "/test/path", storage.DocumentPath)
		assert.Equal(t, int64(10485760), storage.MaxFileSize)
	})
}

func TestDBPoolConfig(t *testing.T) {
	t.Run("db pool config creation", func(t *testing.T) {
		pool := DBPoolConfig{
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 1 * time.Minute,
			ConnectTimeout:  10 * time.Second,
			QueryTimeout:    30 * time.Second,
			MaxRetries:      3,
			RetryDelay:      1 * time.Second,
		}

		assert.Equal(t, 25, pool.MaxOpenConns)
		assert.Equal(t, 10, pool.MaxIdleConns)
		assert.Equal(t, 5*time.Minute, pool.ConnMaxLifetime)
		assert.Equal(t, 1*time.Minute, pool.ConnMaxIdleTime)
		assert.Equal(t, 10*time.Second, pool.ConnectTimeout)
		assert.Equal(t, 30*time.Second, pool.QueryTimeout)
		assert.Equal(t, 3, pool.MaxRetries)
		assert.Equal(t, 1*time.Second, pool.RetryDelay)
	})
}

func TestCoreConfig_Integration(t *testing.T) {
	t.Run("full integration test", func(t *testing.T) {
		// Setup test environment
		envVars := map[string]string{
			"ENVIRONMENT":   constants.EnvDevelopment,
			"DOCUMENT_PATH": "/test/documents",
			"MAX_FILE_SIZE": "20971520", // 20MB
			"SERVER_PORT":   "8080",
			"JWT_SECRET":    "test-jwt-secret",
		}

		for key, value := range envVars {
			_ = os.Setenv(key, value)
		}

		defer func() {
			for key := range envVars {
				_ = os.Unsetenv(key)
			}
		}()

		// Create test .env file
		createTestEnvFile(t, ".env.test")

		defer func() {
			_ = os.Remove(".env.test")
		}()

		// Create config
		cfg, err := NewCoreConfig(".env.test")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Validate config
		err = cfg.Validate()
		assert.NoError(t, err)

		// Check storage config (optional; defaults loaded from DOCUMENT_PATH / MAX_FILE_SIZE)
		assert.Equal(t, "/test/documents", cfg.Storage.DocumentPath)
		assert.Equal(t, int64(20971520), cfg.Storage.MaxFileSize)

		// Check DB pool config (should match development config)
		devConfig := databases.DevelopmentConfig()
		assert.Equal(t, devConfig.MaxOpenConns, cfg.DBPool.MaxOpenConns)
		assert.Equal(t, devConfig.MaxIdleConns, cfg.DBPool.MaxIdleConns)
		assert.Equal(t, devConfig.ConnMaxLifetime, cfg.DBPool.ConnMaxLifetime)
		assert.Equal(t, devConfig.ConnMaxIdleTime, cfg.DBPool.ConnMaxIdleTime)
		assert.Equal(t, devConfig.ConnectTimeout, cfg.DBPool.ConnectTimeout)
		assert.Equal(t, devConfig.QueryTimeout, cfg.DBPool.QueryTimeout)
		assert.Equal(t, devConfig.MaxRetries, cfg.DBPool.MaxRetries)
		assert.Equal(t, devConfig.RetryDelay, cfg.DBPool.RetryDelay)
	})
}

func TestCoreConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		cleanup     func()
		expectError bool
	}{
		{
			name: "very large max file size",
			setup: func() {
				_ = os.Setenv("MAX_FILE_SIZE", "1073741824") // 1GB
			},
			cleanup: func() {
				_ = os.Unsetenv("MAX_FILE_SIZE")
			},
			expectError: false,
		},
		{
			name: "empty document path with spaces",
			setup: func() {
				_ = os.Setenv("DOCUMENT_PATH", "   ")
			},
			cleanup: func() {
				_ = os.Unsetenv("DOCUMENT_PATH")
			},
			expectError: false, // This should be handled by the config package
		},
		{
			name: "invalid max file size string",
			setup: func() {
				_ = os.Setenv("MAX_FILE_SIZE", "invalid")
			},
			cleanup: func() {
				_ = os.Unsetenv("MAX_FILE_SIZE")
			},
			expectError: false, // Should use default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			createTestEnvFile(t, ".env.test")

			defer func() {
				_ = os.Remove(".env.test")
			}()

			cfg, err := NewCoreConfig(".env.test")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

// Helper function to create a test .env file
func createTestEnvFile(t *testing.T, filename string) {
	content := fmt.Sprintf(`# Test environment file
		ENVIRONMENT=%s
		DOCUMENT_PATH=./storage/documents
		MAX_FILE_SIZE=10485760
		SERVER_PORT=8080
		JWT_SECRET=test-secret
	`, constants.EnvDevelopment)
	err := os.WriteFile(filename, []byte(content), 0o600)
	require.NoError(t, err)
}

// Helper function to create a test .env file for benchmarks
func createTestEnvFileForBench(b *testing.B, filename string) {
	content := fmt.Sprintf(`# Test environment file
		ENVIRONMENT=%s
		DOCUMENT_PATH=./storage/documents
		MAX_FILE_SIZE=10485760
		SERVER_PORT=8080
		JWT_SECRET=test-secret
	`, constants.EnvDevelopment)

	err := os.WriteFile(filename, []byte(content), 0o600)
	if err != nil {
		b.Fatal(err)
	}
}

// Benchmark tests
func BenchmarkNewCoreConfig(b *testing.B) {
	// Setup
	createTestEnvFileForBench(b, ".env.bench")

	defer func() {
		_ = os.Remove(".env.bench")
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := NewCoreConfig(".env.bench")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetDBPoolConfig(b *testing.B) {
	_ = os.Setenv("ENVIRONMENT", constants.EnvDevelopment)
	defer func() {
		_ = os.Unsetenv("ENVIRONMENT")
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = getDBPoolConfig()
	}
}

func BenchmarkCoreConfig_Validate(b *testing.B) {
	cfg := &CoreConfig{
		BaseConfig: config.BaseConfig{
			Server: config.ServerConfig{Port: "8080"},
			Auth:   config.AuthConfig{JWTSecret: "test-secret"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}
