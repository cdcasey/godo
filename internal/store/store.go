package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/tursodatabase/go-libsql"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrTodoNotFound = errors.New("todo not found")
)

type Store struct {
	db *sql.DB
}

func New(databaseURL string, authToken string) (*Store, error) {
	connStr := databaseURL
	if authToken != "" {
		connStr = fmt.Sprintf("%s?authToken=%s", databaseURL, authToken)
	}

	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{db: db}, nil

}

func (s *Store) RunMigrations(migratopnPath string) error {
	// Create migrations table if not exists
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	files, err := os.ReadDir(migratopnPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Get .up.sql files and sort them
	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	// Run each migration
	for _, fileName := range upFiles {
		version := strings.TrimSuffix(fileName, ".up.sql")

		// Check is already applied
		var count int
		err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migrations status: %w", err)
		}
		if count > 0 {
			continue
		}

		// Read and execute migrations
		content, err := os.ReadFile(filepath.Join(migratopnPath, fileName))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", fileName, err)
		}

		_, err = s.db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", fileName, err)
		}

		// Record migration
		_, err = s.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", fileName, err)
		}
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
