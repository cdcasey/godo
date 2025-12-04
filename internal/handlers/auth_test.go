package handlers

import (
	"bytes"
	"encoding/json"
	"godo/internal/models"
	"godo/internal/store"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestHandler(t *testing.T) (*AuthHandler, *store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	testStore, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	if err := testStore.RunMigrations("../../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() { testStore.Close() })

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := NewAuthHandler(testStore, logger, "test-jwt-secret")

	return handler, testStore
}

func TestRegister_Success(t *testing.T) {
	handler, _ := setupTestHandler(t)

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Token == "" {
		t.Error("Expected token, got empty string")
	}

	if resp.User.Email != reqBody.Email {
		t.Errorf("Expected email %s, got %s", reqBody.Email, resp.User.Email)
	}

	if resp.User.Role != models.RoleUser {
		t.Errorf("Expected role %s, got %s", models.RoleUser, resp.User.Role)
	}
}

func TestRegister_Failure(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid json",
			body:           "not-json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "missing email",
			body: RegisterRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
		{
			name: "missing password",
			body: RegisterRequest{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
		{
			name: "password too short",
			body: RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := setupTestHandler(t)

			var body []byte
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if !bytes.Contains(rec.Body.Bytes(), []byte(tt.expectedError)) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, rec.Body.String())
			}
		})
	}

}
