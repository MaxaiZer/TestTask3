package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log/slog"
)

func RunMigrations(db *sql.DB, migrationsPath string) error {

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations_wallet",
	})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no new migrations to apply")
		} else {
			return fmt.Errorf("could not apply migrations: %w", err)
		}
	} else {
		slog.Info("migrations were applied successfully")
	}
	return nil
}
