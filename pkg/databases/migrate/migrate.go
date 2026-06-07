package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Options contains the options for the migration
type Options struct {
	DatabaseURL    string
	MigrationsPath string
}

// Up runs all pending migrations
func Up(opts Options) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", opts.MigrationsPath),
		opts.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %v", err)
	}
	return nil
}

// Down reverts all migrations
func Down(opts Options) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", opts.MigrationsPath),
		opts.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error reverting migrations: %v", err)
	}
	return nil
}

// Force forces a specific version
func Force(opts Options, version int) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", opts.MigrationsPath),
		opts.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("error forcing version: %v", err)
	}
	return nil
}

// Version gets the current version
func Version(opts Options) (uint, bool, error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", opts.MigrationsPath),
		opts.DatabaseURL,
	)
	if err != nil {
		return 0, false, fmt.Errorf("error creating migrate instance: %v", err)
	}
	defer m.Close()

	return m.Version()
}
