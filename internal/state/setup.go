package state

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	_ "modernc.org/sqlite"

	"github.com/computer-technology-team/download-manager.git/datadir"
)

const databaseFileName = "sqlite.db"

//go:embed schemas/*.sql
var schemasFS embed.FS

func SetupDatabase(ctx context.Context) (*sql.DB, error) {
	dataDir, err := datadir.GetAppDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine app data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, databaseFileName)

	dsn := fmt.Sprintf("file:%s?_foreign_keys=on", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrateDriver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	fsys := http.FS(schemasFS)

	schemasSource, err := httpfs.New(fsys, "schemas")
	if err != nil {
		return fmt.Errorf("failed to create schema source: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"httpfs", schemasSource,
		"sqlite", migrateDriver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
