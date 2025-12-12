package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"godo/internal/service"
	"godo/internal/store"
	"godo/internal/testutil"
)

func setupWebTestHandler(t *testing.T) *WebHandler {
	t.Helper()

	db := testutil.SetupTestDB(t)
	userRepo := store.NewUserRepo(db)
	authService := service.NewAuthService(userRepo)

	return NewWebHandler(authService, "test-jwt-secret")
}

func TestWebLoginPage_Renders(t *testing.T) {
	handler := setupWebTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	handler.LoginPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<form") {
		t.Error("Expected response to contain a form")
	}
	if !strings.Contains(body, "email") {
		t.Error("Expected response to contain email field")
	}
}

func TestWebLogin_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userRepo := store.NewUserRepo(db)
	authService := service.NewAuthService(userRepo)
	handler := NewWebHandler(authService, "test-jwt-secret")

	// Create a user
	password := "password123"
	_, err := authService.Register("test@example.com", password)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Submit login form
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", password)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	// Check for redirect header
	if rec.Header().Get("HX-Redirect") != "/todos" {
		t.Errorf("Expected HX-Redirect to /todos, got %s", rec.Header().Get("HX-Redirect"))
	}

	// Check for auth cookie
	cookies := rec.Result().Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "auth_token" {
			authCookie = c
			break
		}
	}

	if authCookie == nil {
		t.Fatal("Expected auth_token cookie to be set")
	}

	if !authCookie.HttpOnly {
		t.Error("Expected cookie to be HttpOnly")
	}
}

func TestWebLogin_InvalidCredentials(t *testing.T) {
	handler := setupWebTestHandler(t)

	form := url.Values{}
	form.Add("email", "nonexistent@example.com")
	form.Add("password", "wrongpassword")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	// Should return 200 with error message (HTMX pattern)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Invalid email or password") {
		t.Errorf("Expected error message, got %s", body)
	}

	// Should NOT set cookie
	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "auth_token" {
			t.Error("Should not set auth cookie on failed login")
		}
	}

	// Should NOT redirect
	if rec.Header().Get("HX-Redirect") != "" {
		t.Error("Should not redirect on failed login")
	}
}
