package test_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

// TestConfigurationValidation valida configuraciones de prueba solo con stubs de shared (sin dependencia de auth).
func TestConfigurationValidation(t *testing.T) {
	err := logging.Init(false)
	require.NoError(t, err)

	configStubs := NewStubs()

	testCases := []struct {
		name        string
		configType  string
		expectError bool
	}{
		{
			name:        "valid configuration",
			configType:  "valid",
			expectError: false,
		},
		{
			name:        "development configuration",
			configType:  constants.EnvDevelopment,
			expectError: false,
		},
		{
			name:        "production configuration",
			configType:  constants.EnvProduction,
			expectError: false,
		},
		{
			name:        "minimal configuration",
			configType:  "minimal",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := configStubs.GetTestConfig(tc.configType)

			err := config.Validate()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.NotEmpty(t, config.Server.Port)
				assert.NotEmpty(t, config.Auth.JWTSecret)
				assert.NotEmpty(t, config.Storage.DocumentPath)
				assert.Greater(t, config.Storage.MaxFileSize, int64(0))
			}
		})
	}
}
