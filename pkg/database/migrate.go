package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func RollbackMigrations(dsn string, migrationsPath string, steps int) error {
	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback migrations: %w", err)
	}

	return nil
}
