package store

import (
	"godo/internal/testutil"
	"testing"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	return testutil.SetupTestStore(t)
}

func TestNew_Success(t *testing.T) {
	store := setupTestStore(t)

	if err := store.db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestRunMigrations_Success(t *testing.T) {
	store := setupTestStore(t)

	// Run migrations
	migrationsPath := "../../migrations"
	if err := store.RunMigrations(migrationsPath); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify tables exist
	var tableName string
	err := store.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	if err != nil {
		t.Errorf("Users table not found: %v", err)
	}
}

func TestClose(t *testing.T) {
	store := setupTestStore(t)

	if err := store.Close(); err != nil {
		t.Errorf("expected no error on close, got %v", err)
	}

	// Verify connection is closed
	if err := store.db.Ping(); err == nil {
		t.Error("expected error after close, got nil")
	}
}

func TestRunMigrations_InvalidPath(t *testing.T) {
	store := setupTestStore(t)

	// Try invalid migrations path
	err := store.RunMigrations("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for invalid migrations path, got nil")
	}
}

func TestRunMigrations_Idempodent(t *testing.T) {
	store := setupTestStore(t)

	migrationsPath := "../../migrations"

	if err := store.RunMigrations(migrationsPath); err != nil {
		t.Fatalf("First migration run failed: %v", err)
	}

	if err := store.RunMigrations(migrationsPath); err != nil {
		t.Fatalf("Second migration run failed: %v", err)
	}
}
