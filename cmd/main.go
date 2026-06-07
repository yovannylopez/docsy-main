package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/yovannylopez/docsy-main/cmd/composition"
	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
	"github.com/yovannylopez/docsy-main/pkg/logging"
)

const (
	// shutdownTimeout defines the maximum time to wait for graceful shutdown
	shutdownTimeout = 30 * time.Second
)

func main() {
	// Load configuration.
	// If .env exists in the working directory it is loaded automatically (useful in local).
	// In Docker/K8s the variables are injected directly and the file is not needed.
	envFile := ""
	if _, err := os.Stat(".env"); err == nil {
		envFile = ".env"
	}

	cfg, err := config.NewCoreConfig(envFile)
	if err != nil {
		panic(fmt.Sprintf("Error loading config: %v", err))
	}

	// Initialize logging in production mode when ENVIRONMENT=production.
	productionMode := os.Getenv("ENVIRONMENT") == "production"
	if err := logging.Init(!productionMode); err != nil {
		panic(err)
	}
	defer func() { _ = logging.Sync() }()

	logging.Info("Starting server", logging.WithRequestID("startup"))

	// Create and configure application
	app, err := composition.NewApplication(cfg)
	if err != nil {
		logging.Error("Error creating application", zap.Error(err))
		panic(err)
	}

	// Configure application (tracing, routes, OpenAPI)
	if err := app.Setup(); err != nil {
		logging.Error("Error setting up application", zap.Error(err))
		panic(err)
	}

	// Configure graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Execute server in goroutine
	go func() {
		if err := app.Run(); err != nil {
			logging.Error("Error running server", zap.Error(err))
			quit <- syscall.SIGTERM
		}
	}()

	// Wait for shutdown signal
	<-quit
	logging.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		logging.Error("Error during shutdown", zap.Error(err))
		panic(fmt.Sprintf("Error during shutdown: %v", err))
	}

	logging.Info("Server stopped gracefully")
}
