package store

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) RunMigrations(migratopnPath string) error {
	driver, err := sqlite3.WithInstance(s.db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("Failed to create migration driver: %w", err)
	}

	fileSource, err := (&file.File{}).Open("file://" + migratopnPath)
	if err != nil {
		return fmt.Errorf("Failed to open migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("file", fileSource, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("Failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Failed to run migrations: %w", err)
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
