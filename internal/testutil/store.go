package testutil

import (
	"path/filepath"
	"testing"

	"godo/internal/store"
)

func SetupTestStore(t *testing.T) *store.Store {
	t.Helper()
	dbPath := "file:" + filepath.Join(t.TempDir(), "test.db")

	s, err := store.New(dbPath, "")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	if err := s.RunMigrations("../../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() { s.Close() })
	return s
}
