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

func NewDB(databaseURL string, authToken string) (*sql.DB, error) {
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

	return db, nil
}

func RunMigrations(db *sql.DB, migrationPath string) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	files, err := os.ReadDir(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		version := strings.TrimSuffix(fileName, ".up.sql")

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationPath, fileName))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", fileName, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", fileName, err)
		}

		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", fileName, err)
		}
	}

	return nil
}
