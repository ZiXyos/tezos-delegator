package database

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func RunMigrations(db *sql.DB, fs embed.FS) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	sourceDriver, err := iofs.New(fs, "database/sql")
	if err != nil {
		return fmt.Errorf("could not create source driver: %w", err)
	}

	migration, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %w", err)
	}
	defer migration.Close()

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	return nil
}

func GetMigrationVersion(db *sql.DB, fs embed.FS) (uint, bool, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("could not create postgres driver: %w", err)
	}

	sourceDriver, err := iofs.New(fs, "database/sql")
	if err != nil {
		return 0, false, fmt.Errorf("could not create source driver: %w", err)
	}

	migration, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return 0, false, fmt.Errorf("could not create migration instance: %w", err)
	}
	defer migration.Close()

	return migration.Version()
}
