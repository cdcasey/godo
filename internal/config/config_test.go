package config

import (
	"os"
	"testing"
)

func TestConfigLoadSuccess(t *testing.T) {
	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.DatabaseURL != "/tmp/test.db" {
		t.Errorf("expected DatabaseURL=/tmp/test.db, got %s", cfg.DatabaseURL)
	}

	if cfg.JWTSecret != "test-secret" {
		t.Errorf("expected JWTSecret=test-secret, got %s", cfg.JWTSecret)
	}

	// Test defaults
	if cfg.Port != "8080" {
		t.Errorf("expected default Port=8080, got %s", cfg.Port)
	}
}
