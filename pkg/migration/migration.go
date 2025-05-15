package migration

import (
	"database/sql"
	"log"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB, migrationsPath string) error {
	log.Printf("Running database migrations from: %s", migrationsPath)

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}

	sourceURL := formatSourceURL(absPath)
	log.Printf("Using migrations source URL: %s", sourceURL)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to apply")
			return nil
		}
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func formatSourceURL(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")

	filepath.Base(path)

	if filepath.IsAbs(path) {
		return "file://migrations"
	}

	return "file://" + path
}
