package composition

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	authpolicies "github.com/yovannylopez/docsy-main/internal/auth/domain/policies"
	authContainer "github.com/yovannylopez/docsy-main/internal/auth/infrastructure/container"
	"github.com/yovannylopez/docsy-main/internal/auth/transport/handlers"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
	sharedHandlers "github.com/yovannylopez/docsy-main/internal/shared/transport/handlers"
	usersContainer "github.com/yovannylopez/docsy-main/internal/users/infrastructure/container"
	usersHandlers "github.com/yovannylopez/docsy-main/internal/users/transport/handlers"
	"github.com/yovannylopez/docsy-main/pkg/databases"
	"github.com/yovannylopez/docsy-main/pkg/ratelimit"
)

// Container contains all the dependencies of the application.
// To add a new module: create <ModuleName>Container here and wire in NewContainer.
type Container struct {
	// Base modules (always present)
	AuthContainer  *authContainer.AuthContainer
	UsersContainer *usersContainer.UsersContainer

	// Database
	DB                    *sqlx.DB
	CircuitBreakerWrapper *databases.CircuitBreakerWrapper

	// Shared handlers
	HealthHandler *sharedHandlers.HealthHandler
	UserHandler   *usersHandlers.UsersHandler

	// Rate limiting (auth public routes: login, refresh, logout, mfa/verify)
	AuthRateLimit   echo.MiddlewareFunc
	authRateLimiter *ratelimit.AuthRateLimiter
}

// NewContainer creates and initializes the main container with all the base modules.
func NewContainer(cfg *config.CoreConfig) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	ctx := context.Background()
	db, cbWrapper, err := databases.NewConnectionWithCircuitBreaker(ctx, databases.Config{
		DatabaseURL:     cfg.Database.DatabaseURL,
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.DBPool.MaxOpenConns,
		MaxIdleConns:    cfg.DBPool.MaxIdleConns,
		ConnMaxLifetime: cfg.DBPool.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBPool.ConnMaxIdleTime,
		ConnectTimeout:  cfg.DBPool.ConnectTimeout,
		QueryTimeout:    cfg.DBPool.QueryTimeout,
		MaxRetries:      cfg.DBPool.MaxRetries,
		RetryDelay:      cfg.DBPool.RetryDelay,
		MaxBackoff:      cfg.DBPool.MaxBackoff,
		CircuitBreaker:  cfg.DBPool.CircuitBreaker,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	authRL := ratelimit.NewAuthRateLimiter(cfg.Redis)
	lockout := authpolicies.FailedLoginLockoutPolicy{
		MaxAttempts:  cfg.Auth.LockoutMaxAttempts,
		LockDuration: cfg.Auth.LockoutDuration,
	}
	authC := authContainer.NewAuthContainerWithMFA(db, cfg.Auth.JWTSecret, cfg.LDAP, cfg.MFA, lockout)
	usersC := usersContainer.NewUsersContainer(
		authC.UserRepository,
		authC.GetPasswordHasher(),
		authC.GetAuditRepository(),
	)

	healthHandler := sharedHandlers.NewHealthHandler(db, cbWrapper)
	userHandler := usersC.GetUsersHandler()

	return &Container{
		AuthContainer:         authC,
		UsersContainer:        usersC,
		DB:                    db,
		CircuitBreakerWrapper: cbWrapper,
		HealthHandler:         healthHandler,
		UserHandler:           userHandler,
		AuthRateLimit:         authRL.Middleware,
		authRateLimiter:       authRL,
	}, nil
}

// Close releases all the resources of the container in reverse order of their creation.
func (c *Container) Close() error {
	if c.authRateLimiter != nil {
		if err := c.authRateLimiter.Close(); err != nil {
			return fmt.Errorf("failed to close rate limiter: %w", err)
		}
	}

	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	return nil
}

// CreateAuthHandler returns the authentication handler.
func (c *Container) CreateAuthHandler() *handlers.AuthHandler {
	return c.AuthContainer.GetAuthHandler()
}

// CreateHealthHandler returns the health handler.
func (c *Container) CreateHealthHandler() *sharedHandlers.HealthHandler {
	return c.HealthHandler
}

// CreateUserHandler returns the users handler.
func (c *Container) CreateUserHandler() *usersHandlers.UsersHandler {
	return c.UserHandler
}

// CreateLoginPageHandler returns the server-rendered login page handler.
func (c *Container) CreateLoginPageHandler() *handlers.LoginPageHandler {
	return handlers.NewLoginPageHandler(
		c.AuthContainer.LoginUseCase,
		c.AuthContainer.AuthUseCase,
	)
}
