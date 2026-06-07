package test_utils

import (
	"time"

	sharedConfig "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
	pkgConfig "github.com/yovannylopez/docsy-main/pkg/config"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// Constants for test configurations
const (
	// Database pool constants
	DefaultMaxOpenConns    = 10
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 5
	DefaultConnMaxIdleTime = 5
	DefaultConnectTimeout  = 10
	DefaultQueryTimeout    = 30
	DefaultMaxRetries      = 3
	DefaultRetryDelay      = 1

	// Production database pool constants
	ProductionMaxOpenConns    = 25
	ProductionMaxIdleConns    = 10
	ProductionConnMaxLifetime = 10
	ProductionConnMaxIdleTime = 5
	ProductionConnectTimeout  = 15
	ProductionQueryTimeout    = 60
	ProductionMaxRetries      = 5
	ProductionRetryDelay      = 2

	// Development database pool constants
	DevelopmentMaxOpenConns    = 5
	DevelopmentMaxIdleConns    = 2
	DevelopmentConnMaxLifetime = 2
	DevelopmentConnMaxIdleTime = 30
	DevelopmentConnectTimeout  = 5
	DevelopmentQueryTimeout    = 15
	DevelopmentMaxRetries      = 2
	DevelopmentRetryDelay      = 500

	// Minimal database pool constants
	MinimalMaxOpenConns    = 1
	MinimalMaxIdleConns    = 1
	MinimalConnMaxLifetime = 1
	MinimalConnMaxIdleTime = 30
	MinimalConnectTimeout  = 5
	MinimalQueryTimeout    = 10
	MinimalMaxRetries      = 1
	MinimalRetryDelay      = 100

	// Server timeout constants
	DefaultReadTimeout  = 30
	DefaultWriteTimeout = 30
	MinimalReadTimeout  = 10
	MinimalWriteTimeout = 10

	// Storage constants
	DefaultMaxFileSize = 10485760 // 10MB in bytes

	// String constants for test types
	TestValidType   = "valid"
	TestInvalidType = "invalid"
	TestEmptyType   = "empty"
)

// TestConfigs holds predefined test configurations
type TestConfigs struct {
	ValidCoreConfig   *sharedConfig.CoreConfig
	InvalidCoreConfig *sharedConfig.CoreConfig
	DevelopmentConfig *sharedConfig.CoreConfig
	ProductionConfig  *sharedConfig.CoreConfig
	MinimalConfig     *sharedConfig.CoreConfig
	DatabaseConfig    pkgConfig.DatabaseConfig
	ServerConfig      pkgConfig.ServerConfig
	AuthConfig        pkgConfig.AuthConfig
	DBPoolConfig      sharedConfig.DBPoolConfig
	StorageConfig     sharedConfig.StorageConfig
}

// Stubs holds only configuration stubs (pure Shared Kernel, without auth/users)
type Stubs struct {
	Configs TestConfigs
}

// NewStubs creates a new stubs instance with only test configurations
func NewStubs() *Stubs {
	return &Stubs{
		Configs: newTestConfigs(),
	}
}

// GetTestConfig returns a specific test configuration
func (s *Stubs) GetTestConfig(configType string) *sharedConfig.CoreConfig {
	switch configType {
	case TestValidType:
		return s.Configs.ValidCoreConfig
	case TestInvalidType:
		return s.Configs.InvalidCoreConfig
	case constants.EnvDevelopment:
		return s.Configs.DevelopmentConfig
	case constants.EnvProduction:
		return s.Configs.ProductionConfig
	case "minimal":
		return s.Configs.MinimalConfig
	default:
		return s.Configs.ValidCoreConfig
	}
}

// newTestConfigs creates test configurations
func newTestConfigs() TestConfigs {
	dbConfig := pkgConfig.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "test_user",
		Password: "test_password",
		DBName:   "test_db",
		SSLMode:  "disable",
	}

	serverConfig := pkgConfig.ServerConfig{
		Host:         "localhost",
		Port:         "8080",
		ReadTimeout:  DefaultReadTimeout * time.Second,
		WriteTimeout: DefaultWriteTimeout * time.Second,
	}

	authConfig := pkgConfig.AuthConfig{
		JWTSecret: "test-jwt-secret-key-32-bytes-long",
	}

	dbPoolConfig := sharedConfig.DBPoolConfig{
		MaxOpenConns:    DefaultMaxOpenConns,
		MaxIdleConns:    DefaultMaxIdleConns,
		ConnMaxLifetime: DefaultConnMaxLifetime * time.Minute,
		ConnMaxIdleTime: DefaultConnMaxIdleTime * time.Minute,
		ConnectTimeout:  DefaultConnectTimeout * time.Second,
		QueryTimeout:    DefaultQueryTimeout * time.Second,
		MaxRetries:      DefaultMaxRetries,
		RetryDelay:      DefaultRetryDelay * time.Second,
	}

	storageConfig := sharedConfig.StorageConfig{
		DocumentPath: "/test/documents",
		MaxFileSize:  DefaultMaxFileSize,
	}

	return TestConfigs{
		ValidCoreConfig: &sharedConfig.CoreConfig{
			BaseConfig: pkgConfig.BaseConfig{
				Server:   serverConfig,
				Database: dbConfig,
				Auth:     authConfig,
			},
			DBPool:  dbPoolConfig,
			Storage: storageConfig,
		},
		InvalidCoreConfig: &sharedConfig.CoreConfig{
			BaseConfig: pkgConfig.BaseConfig{
				Server: pkgConfig.ServerConfig{
					Host:         "invalid-host",
					Port:         "8080",
					ReadTimeout:  DefaultReadTimeout * time.Second,
					WriteTimeout: DefaultWriteTimeout * time.Second,
				},
				Database: pkgConfig.DatabaseConfig{
					Host:     "invalid-host",
					Port:     "5432",
					User:     "test_user",
					Password: "test_password",
					DBName:   "test_db",
					SSLMode:  "disable",
				},
				Auth: pkgConfig.AuthConfig{
					JWTSecret: "",
				},
			},
			DBPool: sharedConfig.DBPoolConfig{
				MaxOpenConns:    DefaultMaxOpenConns,
				MaxIdleConns:    DefaultMaxIdleConns,
				ConnMaxLifetime: DefaultConnMaxLifetime * time.Minute,
				ConnMaxIdleTime: DefaultConnMaxIdleTime * time.Minute,
				ConnectTimeout:  1 * time.Second,
				QueryTimeout:    DefaultQueryTimeout * time.Second,
				MaxRetries:      MinimalMaxRetries,
				RetryDelay:      MinimalRetryDelay * time.Millisecond,
			},
			Storage: storageConfig,
		},
		DevelopmentConfig: &sharedConfig.CoreConfig{
			BaseConfig: pkgConfig.BaseConfig{
				Server:   serverConfig,
				Database: dbConfig,
				Auth:     authConfig,
			},
			DBPool: sharedConfig.DBPoolConfig{
				MaxOpenConns:    DevelopmentMaxOpenConns,
				MaxIdleConns:    DevelopmentMaxIdleConns,
				ConnMaxLifetime: DevelopmentConnMaxLifetime * time.Minute,
				ConnMaxIdleTime: DevelopmentConnMaxIdleTime * time.Second,
				ConnectTimeout:  DevelopmentConnectTimeout * time.Second,
				QueryTimeout:    DevelopmentQueryTimeout * time.Second,
				MaxRetries:      DevelopmentMaxRetries,
				RetryDelay:      DevelopmentRetryDelay * time.Millisecond,
			},
			Storage: storageConfig,
		},
		ProductionConfig: &sharedConfig.CoreConfig{
			BaseConfig: pkgConfig.BaseConfig{
				Server:   serverConfig,
				Database: dbConfig,
				Auth:     authConfig,
			},
			DBPool: sharedConfig.DBPoolConfig{
				MaxOpenConns:    ProductionMaxOpenConns,
				MaxIdleConns:    ProductionMaxIdleConns,
				ConnMaxLifetime: ProductionConnMaxLifetime * time.Minute,
				ConnMaxIdleTime: ProductionConnMaxIdleTime * time.Minute,
				ConnectTimeout:  ProductionConnectTimeout * time.Second,
				QueryTimeout:    ProductionQueryTimeout * time.Second,
				MaxRetries:      ProductionMaxRetries,
				RetryDelay:      ProductionRetryDelay * time.Second,
			},
			Storage: storageConfig,
		},
		MinimalConfig: &sharedConfig.CoreConfig{
			BaseConfig: pkgConfig.BaseConfig{
				Server: pkgConfig.ServerConfig{
					Host:         "localhost",
					Port:         "8080",
					ReadTimeout:  MinimalReadTimeout * time.Second,
					WriteTimeout: MinimalWriteTimeout * time.Second,
				},
				Database: dbConfig,
				Auth:     authConfig,
			},
			DBPool: sharedConfig.DBPoolConfig{
				MaxOpenConns:    MinimalMaxOpenConns,
				MaxIdleConns:    MinimalMaxIdleConns,
				ConnMaxLifetime: MinimalConnMaxLifetime * time.Minute,
				ConnMaxIdleTime: MinimalConnMaxIdleTime * time.Second,
				ConnectTimeout:  MinimalConnectTimeout * time.Second,
				QueryTimeout:    MinimalQueryTimeout * time.Second,
				MaxRetries:      MinimalMaxRetries,
				RetryDelay:      MinimalRetryDelay * time.Millisecond,
			},
			Storage: storageConfig,
		},
		DatabaseConfig: dbConfig,
		ServerConfig:   serverConfig,
		AuthConfig:     authConfig,
		DBPoolConfig:   dbPoolConfig,
		StorageConfig:  storageConfig,
	}
}
