package config

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yovannylopez/docsy-main/pkg/config/mocks"
)

// TestAppConfig is a test implementation of AppConfig
type TestAppConfig struct {
	validationError error
}

func (t *TestAppConfig) Validate() error {
	return t.validationError
}

func TestLoadConfig_Success(t *testing.T) {
	// Arrange
	cfg := &TestAppConfig{validationError: nil}

	// Act
	result, err := LoadConfig("", cfg)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, cfg, result)
}

func TestLoadConfig_ValidationError(t *testing.T) {
	// Arrange
	expectedError := errors.New("validation failed")
	cfg := &TestAppConfig{validationError: expectedError}

	// Act
	result, err := LoadConfig("", cfg)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config validation error")
	assert.Contains(t, err.Error(), expectedError.Error())
	assert.Equal(t, cfg, result)
}

func TestLoadConfig_WithEnvFile(t *testing.T) {
	// Create a temporary .env file
	envContent := "TEST_VAR=test_value"
	tmpFile, err := os.CreateTemp("", "test.env")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(envContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Arrange
	cfg := &TestAppConfig{validationError: nil}

	// Act
	result, err := LoadConfig(tmpFile.Name(), cfg)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, cfg, result)
}

func TestLoadConfig_InvalidEnvFile(t *testing.T) {
	// Arrange
	cfg := &TestAppConfig{validationError: nil}

	// Act
	result, err := LoadConfig("nonexistent.env", cfg)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error loading env file")
	assert.Equal(t, cfg, result)
}

func TestGetEnv_WithValue(t *testing.T) {
	// Arrange
	key := "TEST_ENV_VAR"
	expectedValue := "test_value"
	os.Setenv(key, expectedValue)
	defer os.Unsetenv(key)

	// Act
	result := GetEnv(key, "default")

	// Assert
	assert.Equal(t, expectedValue, result)
}

func TestGetEnv_WithoutValue(t *testing.T) {
	// Arrange
	key := "NONEXISTENT_ENV_VAR"
	defaultValue := "default_value"

	// Act
	result := GetEnv(key, defaultValue)

	// Assert
	assert.Equal(t, defaultValue, result)
}

func TestGetIntEnv_WithValidValue(t *testing.T) {
	// Arrange
	key := "TEST_INT_VAR"
	expectedValue := 42
	os.Setenv(key, "42")
	defer os.Unsetenv(key)

	// Act
	result := GetIntEnv(key, 0)

	// Assert
	assert.Equal(t, expectedValue, result)
}

func TestGetIntEnv_WithInvalidValue(t *testing.T) {
	// Arrange
	key := "TEST_INVALID_INT_VAR"
	defaultValue := 100
	os.Setenv(key, "not_a_number")
	defer os.Unsetenv(key)

	// Act
	result := GetIntEnv(key, defaultValue)

	// Assert
	assert.Equal(t, defaultValue, result)
}

func TestGetIntEnv_WithoutValue(t *testing.T) {
	// Arrange
	key := "NONEXISTENT_INT_VAR"
	defaultValue := 200

	// Act
	result := GetIntEnv(key, defaultValue)

	// Assert
	assert.Equal(t, defaultValue, result)
}

func TestGetInt64Env_WithValidValue(t *testing.T) {
	// Arrange
	key := "TEST_INT64_VAR"
	expectedValue := int64(9223372036854775807)
	os.Setenv(key, "9223372036854775807")
	defer os.Unsetenv(key)

	// Act
	result := GetInt64Env(key, 0)

	// Assert
	assert.Equal(t, expectedValue, result)
}

func TestGetInt64Env_WithInvalidValue(t *testing.T) {
	// Arrange
	key := "TEST_INVALID_INT64_VAR"
	defaultValue := int64(300)
	os.Setenv(key, "not_a_number")
	defer os.Unsetenv(key)

	// Act
	result := GetInt64Env(key, defaultValue)

	// Assert
	assert.Equal(t, defaultValue, result)
}

func TestGetBoolEnv_WithTrueValue(t *testing.T) {
	// Arrange
	key := "TEST_BOOL_VAR"
	os.Setenv(key, "true")
	defer os.Unsetenv(key)

	// Act
	result := GetBoolEnv(key, false)

	// Assert
	assert.True(t, result)
}

func TestGetBoolEnv_WithFalseValue(t *testing.T) {
	// Arrange
	key := "TEST_BOOL_VAR"
	os.Setenv(key, "false")
	defer os.Unsetenv(key)

	// Act
	result := GetBoolEnv(key, true)

	// Assert
	assert.False(t, result)
}

func TestGetBoolEnv_WithInvalidValue(t *testing.T) {
	// Arrange
	key := "TEST_INVALID_BOOL_VAR"
	defaultValue := true
	os.Setenv(key, "not_a_bool")
	defer os.Unsetenv(key)

	// Act
	result := GetBoolEnv(key, defaultValue)

	// Assert
	assert.Equal(t, defaultValue, result)
}

func TestGetDurationEnv_WithValidValue(t *testing.T) {
	// Arrange
	key := "TEST_DURATION_VAR"
	expectedValue := 30 * time.Second
	os.Setenv(key, "30s")
	defer os.Unsetenv(key)

	// Act
	result := GetDurationEnv(key, time.Hour)

	// Assert
	assert.Equal(t, expectedValue, result)
}

func TestGetDurationEnv_WithInvalidValue(t *testing.T) {
	// Arrange
	key := "TEST_INVALID_DURATION_VAR"
	defaultValue := time.Hour
	os.Setenv(key, "not_a_duration")
	defer os.Unsetenv(key)

	// Act
	result := GetDurationEnv(key, defaultValue)

	// Assert
	assert.Equal(t, defaultValue, result)
}

func TestGetBaseConfig_DefaultValues(t *testing.T) {
	// Act
	config := GetBaseConfig()

	// Assert
	assert.Equal(t, "8100", config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 10*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, config.Server.WriteTimeout)

	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, "5432", config.Database.Port)
	assert.Equal(t, "postgres", config.Database.User)
	assert.Equal(t, "", config.Database.Password)
	assert.Equal(t, "go_boilerplate", config.Database.DBName)
	assert.Equal(t, "disable", config.Database.SSLMode)

	assert.Equal(t, "", config.Auth.JWTSecret)
	assert.Equal(t, 24*time.Hour, config.Auth.TokenExpiration)
	assert.Equal(t, 7*24*time.Hour, config.Auth.RefreshDuration)
	assert.Equal(t, 3, config.Auth.LockoutMaxAttempts)
	assert.Equal(t, 15*time.Minute, config.Auth.LockoutDuration)

	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "stdout", config.Logging.OutputPath)
	assert.Equal(t, "json", config.Logging.Format)
}

func TestGetBaseConfig_WithEnvironmentVariables(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"SERVER_PORT":          "8080",
		"SERVER_HOST":          "127.0.0.1",
		"SERVER_READ_TIMEOUT":  "5s",
		"SERVER_WRITE_TIMEOUT": "5s",
		"DB_HOST":              "db.example.com",
		"DB_PORT":              "5433",
		"DB_USER":              "testuser",
		"DB_PASSWORD":          "testpass",
		"DB_NAME":              "testdb",
		"DB_SSLMODE":           "require",
		"JWT_SECRET":           "test-secret",
		"TOKEN_EXPIRATION":     "1h",
		"REFRESH_DURATION":     "24h",
		"LOG_LEVEL":            "debug",
		"LOG_OUTPUT":           "file.log",
		"LOG_FORMAT":           "text",
	}

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	// Act
	config := GetBaseConfig()

	// Assert
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, 5*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 5*time.Second, config.Server.WriteTimeout)

	assert.Equal(t, "db.example.com", config.Database.Host)
	assert.Equal(t, "5433", config.Database.Port)
	assert.Equal(t, "testuser", config.Database.User)
	assert.Equal(t, "testpass", config.Database.Password)
	assert.Equal(t, "testdb", config.Database.DBName)
	assert.Equal(t, "require", config.Database.SSLMode)

	assert.Equal(t, "test-secret", config.Auth.JWTSecret)
	assert.Equal(t, time.Hour, config.Auth.TokenExpiration)
	assert.Equal(t, 24*time.Hour, config.Auth.RefreshDuration)

	assert.Equal(t, "debug", config.Logging.Level)
	assert.Equal(t, "file.log", config.Logging.OutputPath)
	assert.Equal(t, "text", config.Logging.Format)
}

// Tests usando mocks
func TestLoadConfig_WithMock(t *testing.T) {
	// Arrange
	mockConfig := mocks.NewAppConfig(t)
	mockConfig.EXPECT().Validate().Return(nil)

	// Act
	result, err := LoadConfig("", mockConfig)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mockConfig, result)
}

func TestLoadConfig_WithMockValidationError(t *testing.T) {
	// Arrange
	mockConfig := mocks.NewAppConfig(t)
	expectedError := errors.New("mock validation failed")
	mockConfig.EXPECT().Validate().Return(expectedError)

	// Act
	result, err := LoadConfig("", mockConfig)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config validation error")
	assert.Contains(t, err.Error(), expectedError.Error())
	assert.Equal(t, mockConfig, result)
}

func TestLoadConfig_WithMockAndEnvFile(t *testing.T) {
	// Crear un archivo .env temporal
	envContent := "TEST_VAR=test_value"
	tmpFile, err := os.CreateTemp("", "test.env")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(envContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Arrange
	mockConfig := mocks.NewAppConfig(t)
	mockConfig.EXPECT().Validate().Return(nil)

	// Act
	result, err := LoadConfig(tmpFile.Name(), mockConfig)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mockConfig, result)
}

// Tests of integration
func TestConfigIntegration(t *testing.T) {
	// Test that combines multiple functions
	envVars := map[string]string{
		"SERVER_PORT": "9090",
		"DB_HOST":     "integration-test.com",
		"LOG_LEVEL":   "warn",
	}

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	// Test GetBaseConfig with environment variables
	config := GetBaseConfig()
	assert.Equal(t, "9090", config.Server.Port)
	assert.Equal(t, "integration-test.com", config.Database.Host)
	assert.Equal(t, "warn", config.Logging.Level)

	// Test LoadConfig with mock
	mockConfig := mocks.NewAppConfig(t)
	mockConfig.EXPECT().Validate().Return(nil)

	result, err := LoadConfig("", mockConfig)
	assert.NoError(t, err)
	assert.Equal(t, mockConfig, result)
}

// Tests of edge cases
func TestGetEnv_EmptyValue(t *testing.T) {
	// Arrange
	key := "EMPTY_ENV_VAR"
	os.Setenv(key, "")
	defer os.Unsetenv(key)

	// Act
	result := GetEnv(key, "default")

	// Assert
	assert.Equal(t, "default", result)
}

func TestGetIntEnv_ZeroValue(t *testing.T) {
	// Arrange
	key := "ZERO_INT_VAR"
	os.Setenv(key, "0")
	defer os.Unsetenv(key)

	// Act
	result := GetIntEnv(key, 100)

	// Assert
	assert.Equal(t, 0, result)
}

func TestGetDurationEnv_ComplexDuration(t *testing.T) {
	// Arrange
	key := "COMPLEX_DURATION_VAR"
	expectedValue := 2*time.Hour + 30*time.Minute + 15*time.Second
	os.Setenv(key, "2h30m15s")
	defer os.Unsetenv(key)

	// Act
	result := GetDurationEnv(key, time.Hour)

	// Assert
	assert.Equal(t, expectedValue, result)
}
