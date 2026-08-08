package composition

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/migrations"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/templates"
	"github.com/yovannylopez/docsy-main/pkg/logging"
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

const errMigrationsNotFound = "no migrations folder 'migrations/core' found. " +
	"Configure the MIGRATIONS_PATH environment variable or run from the project root"

// Application represents the composed application with all modules
type Application struct {
	echo      *echo.Echo
	container *Container
	config    *config.CoreConfig
	openapi   *openapi.Generator
}

// NewApplication creates a new instance of the application
func NewApplication(cfg *config.CoreConfig) (*Application, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	container, err := NewContainer(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating container: %w", err)
	}

	e := echo.New()
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	openapiGen := openapi.NewGenerator(
		"Core",
		"Docsy — API base ready for production",
		"1.0.0",
	)

	serverURL := fmt.Sprintf("http://%s:%s", cfg.Server.Host, cfg.Server.Port)
	openapiGen.AddServer(serverURL, "Development server")
	openapiGen.AddTag("authentication", "Authentication operations")
	openapiGen.AddTag("mfa", "Multi-Factor Authentication (TOTP)")
	openapiGen.AddTag("users", "Users operations")
	openapiGen.AddTag("archive", "Personal archive (documents and workspaces)")
	// Add here the OpenAPI tags of your business modules:
	// openapiGen.AddTag("products", "Product management")

	return &Application{
		echo:      e,
		container: container,
		config:    cfg,
		openapi:   openapiGen,
	}, nil
}

// Setup configures the application (routes, OpenAPI).
func (a *Application) Setup() error {
	renderer, err := templates.NewRenderer()
	if err != nil {
		return fmt.Errorf("error loading HTML templates: %w", err)
	}
	a.echo.Renderer = renderer

	router := NewRouter(a.echo, a.container)
	router.SetupRoutes()

	openapi.SetupOpenAPIRoutes(a.echo, a.openapi)

	if err := a.openapi.GenerateFromEcho(a.echo); err != nil {
		logging.Error("Error generating OpenAPI spec", zap.Error(err))
	}

	SetupAllSpecs(a.openapi)

	return nil
}

// Run executes the migrations and starts the server
func (a *Application) Run() error {
	if err := a.runMigrations(); err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	serverAddr := fmt.Sprintf(":%s", a.config.Server.Port)
	logging.Info("Server started", logging.WithRequestID("startup"), zap.String("port", a.config.Server.Port))

	return errors.Wrap(a.echo.Start(serverAddr), "failed to start server")
}

// Shutdown closes the server and releases all resources.
func (a *Application) Shutdown(ctx context.Context) error {
	if err := a.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	if a.container != nil {
		if err := a.container.Close(); err != nil {
			return fmt.Errorf("error closing container: %w", err)
		}
	}

	return nil
}

func (a *Application) runMigrations() error {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		a.config.Database.User,
		a.config.Database.Password,
		a.config.Database.Host,
		a.config.Database.Port,
		a.config.Database.DBName,
		a.config.Database.SSLMode)

	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		return fmt.Errorf("error getting migrations path: %w", err)
	}

	var migrationSuccess bool
	defer func() {
		if !migrationSuccess {
			logging.Warn("Starting rollback of migrations due to an error in the service startup")

			if rollbackErr := migrations.Rollback(dbURL, migrationsPath); rollbackErr != nil {
				logging.Error("Error during the rollback of migrations", zap.Error(rollbackErr))
			} else {
				logging.Info("Rollback of migrations completed successfully")
			}
		}
	}()

	if err := migrations.Setup(dbURL, migrationsPath); err != nil {
		logging.Error("Error executing migrations", zap.Error(err))

		return errors.Wrap(err, "failed to setup database migrations")
	}

	migrationSuccess = true

	logging.Info("Database migrations executed successfully")

	return nil
}

func resolveMigrationsPath() (string, error) {
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		if abs, err := filepath.Abs(envPath); err == nil {
			if info, err2 := os.Stat(abs); err2 == nil && info.IsDir() {
				return abs, nil
			}
		}
	}

	candidates := []string{
		"migrations/core",
		"../migrations/core",
		"../../migrations/core",
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if info, err2 := os.Stat(abs); err2 == nil && info.IsDir() {
				return abs, nil
			}
		}
	}

	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for i := 0; i < 5; i++ {
			candidate := filepath.Join(dir, "migrations", "core")
			if info, err2 := os.Stat(candidate); err2 == nil && info.IsDir() {
				if abs, err3 := filepath.Abs(candidate); err3 == nil {
					return abs, nil
				}

				return candidate, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return "", fmt.Errorf("%s", errMigrationsNotFound)
}
