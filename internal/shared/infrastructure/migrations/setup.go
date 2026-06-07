package migrations

import (
	"fmt"

	"github.com/yovannylopez/docsy-main/pkg/databases/migrate"
)

// runUp runs migrations using the provided function (injectable for tests)
func runUp(dbURL, migrationsPath string, up func(migrate.Options) error) error {
	opts := migrate.Options{
		DatabaseURL:    dbURL,
		MigrationsPath: migrationsPath,
	}
	if err := up(opts); err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}
	return nil
}

// runDown rolls back migrations using the provided function (injectable for tests)
func runDown(dbURL, migrationsPath string, down func(migrate.Options) error) error {
	opts := migrate.Options{
		DatabaseURL:    dbURL,
		MigrationsPath: migrationsPath,
	}
	if err := down(opts); err != nil {
		return fmt.Errorf("error running migration rollback: %w", err)
	}
	return nil
}

// Setup runs the necessary migrations for the Core
func Setup(dbURL, migrationsPath string) error {
	return runUp(dbURL, migrationsPath, migrate.Up)
}

// Rollback reverts all applied migrations
func Rollback(dbURL, migrationsPath string) error {
	return runDown(dbURL, migrationsPath, migrate.Down)
}
