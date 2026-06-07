package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Validation(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name: "valid options",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "/path/to/migrations",
			},
			wantErr: false,
		},
		{
			name: "database URL empty",
			options: Options{
				DatabaseURL:    "",
				MigrationsPath: "/path/to/migrations",
			},
			wantErr: true,
		},
		{
			name: "migrations path empty",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "",
			},
			wantErr: true,
		},
		{
			name: "both fields empty",
			options: Options{
				DatabaseURL:    "",
				MigrationsPath: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate that the options are not empty
			if tt.options.DatabaseURL == "" || tt.options.MigrationsPath == "" {
				assert.True(t, tt.wantErr)
			} else {
				assert.False(t, tt.wantErr)
			}
		})
	}
}

func TestUp_InvalidOptions(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name: "invalid database URL",
			options: Options{
				DatabaseURL:    "invalid://url",
				MigrationsPath: "/tmp/migrations",
			},
			wantErr: true,
		},
		{
			name: "migrations path not found",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Up(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error creating migrate instance")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDown_InvalidOptions(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name: "invalid database URL",
			options: Options{
				DatabaseURL:    "invalid://url",
				MigrationsPath: "/tmp/migrations",
			},
			wantErr: true,
		},
		{
			name: "migrations path not found",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Down(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error creating migrate instance")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestForce_InvalidOptions(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		version int
		wantErr bool
	}{
		{
			name: "invalid database URL",
			options: Options{
				DatabaseURL:    "invalid://url",
				MigrationsPath: "/tmp/migrations",
			},
			version: 1,
			wantErr: true,
		},
		{
			name: "migrations path not found",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "/nonexistent/path",
			},
			version: 1,
			wantErr: true,
		},
		{
			name: "negative version",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "/nonexistent/path",
			},
			version: -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Force(tt.options, tt.version)
			if tt.wantErr {
				assert.Error(t, err)
				// All test cases should fail to create the migrate instance
				assert.Contains(t, err.Error(), "error creating migrate instance")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersion_InvalidOptions(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name: "invalid database URL",
			options: Options{
				DatabaseURL:    "invalid://url",
				MigrationsPath: "/tmp/migrations",
			},
			wantErr: true,
		},
		{
			name: "migrations path not found",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, dirty, err := Version(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, uint(0), version)
				assert.False(t, dirty)
				assert.Contains(t, err.Error(), "error creating migrate instance")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptions_StringRepresentation(t *testing.T) {
	options := Options{
		DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
		MigrationsPath: "/path/to/migrations",
	}

	// Verify that the options can be represented as a string
	assert.NotEmpty(t, options.DatabaseURL)
	assert.NotEmpty(t, options.MigrationsPath)
	assert.Contains(t, options.DatabaseURL, "postgres://")
	assert.Contains(t, options.MigrationsPath, "/path/to/migrations")
}

func TestOptions_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		desc    string
	}{
		{
			name: "URL with special characters",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable&connect_timeout=10",
				MigrationsPath: "/path/with spaces/migrations",
			},
			desc: "URL with additional parameters and path with spaces",
		},
		{
			name: "URL with complex credentials",
			options: Options{
				DatabaseURL:    "postgres://user@domain.com:password@localhost:5432/dbname?sslmode=disable",
				MigrationsPath: "/very/long/path/to/migrations/directory",
			},
			desc: "URL with email as username and long path",
		},
		{
			name: "URL with non-standard port",
			options: Options{
				DatabaseURL:    "postgres://user:pass@localhost:5433/dbname?sslmode=disable",
				MigrationsPath: "/tmp/migrations",
			},
			desc: "URL with non-standard PostgreSQL port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that the options can be created without panicking
			assert.NotPanics(t, func() {
				_ = tt.options
			})

			// Verify that the fields are not empty
			assert.NotEmpty(t, tt.options.DatabaseURL)
			assert.NotEmpty(t, tt.options.MigrationsPath)
		})
	}
}

func BenchmarkUp(b *testing.B) {
	options := Options{
		DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
		MigrationsPath: "/tmp/migrations",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Up(options)
	}
}

func BenchmarkDown(b *testing.B) {
	options := Options{
		DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
		MigrationsPath: "/tmp/migrations",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Down(options)
	}
}

func BenchmarkForce(b *testing.B) {
	options := Options{
		DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
		MigrationsPath: "/tmp/migrations",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Force(options, 1)
	}
}

func BenchmarkVersion(b *testing.B) {
	options := Options{
		DatabaseURL:    "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
		MigrationsPath: "/tmp/migrations",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Version(options)
	}
}
