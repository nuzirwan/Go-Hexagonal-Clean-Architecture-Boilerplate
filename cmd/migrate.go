package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migrations",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE:  runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback last migration",
	RunE:  runMigrateDown,
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	slog.Info("migrations applied successfully")
	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate down: %w", err)
	}
	slog.Info("migration rolled back successfully")
	return nil
}

func newMigrate() (*migrate.Migrate, error) {
	migrationsPath := "file://migrations"

	// Allow override via env
	if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
		migrationsPath = p
	}

	m, err := migrate.New(migrationsPath, cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("create migrate instance: %w", err)
	}
	return m, nil
}
