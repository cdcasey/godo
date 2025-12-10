package testutil

import (
	"database/sql"
	"path/filepath"
	"testing"

	"godo/internal/store"
)

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db")

	db, err := store.NewDB(dbPath, "")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := store.RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}
