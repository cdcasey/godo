package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db")

	db, err := NewDB(dbPath, "")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewDB_Success(t *testing.T) {
	db := setupTestDB(t)

	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestRunMigrations_Success(t *testing.T) {
	db := setupTestDB(t)

	// Verify tables exist
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	if err != nil {
		t.Errorf("Users table not found: %v", err)
	}
}

func TestRunMigrations_InvalidPath(t *testing.T) {
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db")
	db, err := NewDB(dbPath, "")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	err = RunMigrations(db, "/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for invalid migrations path, got nil")
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db := setupTestDB(t)

	// Running migrations again should not fail
	if err := RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("Second migration run failed: %v", err)
	}
}
