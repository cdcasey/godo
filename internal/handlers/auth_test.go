package handlers

import (
	"bytes"
	"encoding/json"
	"godo/internal/models"
	"godo/internal/store"
	"godo/internal/testutil"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func setupTestHandler(t *testing.T) (*AuthHandler, *store.Store) {
	t.Helper()

	testStore := testutil.SetupTestStore(t)

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

func TestLogin_Success(t *testing.T) {
	handler, testStore := setupTestHandler(t)

	// Create a user first
	password := "password123"
	hashedPassword, _ := models.HashPassword(password)
	user := &models.User{
		ID:           models.NewID(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         models.RoleUser,
		CreatedAt:    time.Now(),
	}

	if err := testStore.CreateUser(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Token == "" {
		t.Error("Expected token, got empty string")
	}

	if resp.User.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, resp.User.Email)
	}
}

func TestLogin_Failures(t *testing.T) {
	tests := []struct {
		name           string
		setupUser      bool
		email          string
		password       string
		loginEmail     string
		loginPassword  string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "user not found",
			setupUser:      false,
			loginEmail:     "nonexistent@example.com",
			loginPassword:  "password123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid username or password",
		},
		{
			name:           "wrong password",
			setupUser:      true,
			email:          "test@example.com",
			password:       "correctpassword",
			loginEmail:     "test@example.com",
			loginPassword:  "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid username or password",
		},
		{
			name:           "missing email",
			setupUser:      false,
			loginEmail:     "",
			loginPassword:  "password123",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
		{
			name:           "missing password",
			setupUser:      false,
			loginEmail:     "test@example.com",
			loginPassword:  "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, testStore := setupTestHandler(t)

			// Setup user if needed
			if tt.setupUser {
				hashedPassword, _ := models.HashPassword(tt.password)
				user := &models.User{
					ID:           models.NewID(),
					Email:        tt.email,
					PasswordHash: hashedPassword,
					Role:         models.RoleUser,
					CreatedAt:    time.Now(),
				}
				if err := testStore.CreateUser(user); err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
			}

			reqBody := LoginRequest{
				Email:    tt.loginEmail,
				Password: tt.loginPassword,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if !bytes.Contains(rec.Body.Bytes(), []byte(tt.expectedError)) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, rec.Body.String())
			}
		})
	}
}
