package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"godo/internal/auth"
	"godo/internal/models"
	"godo/internal/store"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func setupTodoTestHandler(t *testing.T) (*TodoHandler, *store.Store) {
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

	handler := NewTodoHandler(testStore, logger)

	return handler, testStore
}

func createTestUser(t *testing.T, s *store.Store, role string) *models.User {
	t.Helper()
	user := &models.User{
		ID:           models.NewID(),
		Email:        "test-" + models.NewID() + "@example.com",
		PasswordHash: "hashed",
		Role:         role,
		CreatedAt:    time.Now(),
	}
	if err := s.CreateUser(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createTestTodo(t *testing.T, s *store.Store, userID string) *models.Todo {
	t.Helper()
	todo := models.NewTodo(userID, "Test Todo", "Test Description")
	if err := s.CreateTodo(todo); err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}
	return todo
}

func requestWithClaims(req *http.Request, claims *auth.Claims) *http.Request {
	ctx := auth.SetClaims(req.Context(), claims)
	return req.WithContext(ctx)
}

func requestWithClaimsAndID(req *http.Request, claims *auth.Claims, paramName, paramValue string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add(paramName, paramValue)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	return requestWithClaims(req, claims)
}

func TestCreate_Success(t *testing.T) {
	handler, _ := setupTodoTestHandler(t)

	user := &auth.Claims{
		UserID: models.NewID(),
		Email:  "test@example.com",
		Role:   models.RoleUser,
	}

	reqBody := CreateTodoRequest{
		Title:       "My Todo",
		Description: "My Description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaims(req, user)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp TodoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Todo.Title != reqBody.Title {
		t.Errorf("Expected title %s, got %s", reqBody.Title, resp.Todo.Title)
	}

	if resp.Todo.UserID != user.UserID {
		t.Errorf("Expected user_id %s, got %s", user.UserID, resp.Todo.UserID)
	}

	if resp.Todo.Completed {
		t.Error("Expected completed to be false")
	}
}

func TestCreate_Failures(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		withClaims     bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "no auth",
			body:           CreateTodoRequest{Title: "Test"},
			withClaims:     false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name:           "invalid json",
			body:           "not-json",
			withClaims:     true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "missing title",
			body:           CreateTodoRequest{Description: "desc"},
			withClaims:     true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := setupTodoTestHandler(t)

			var body []byte
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.withClaims {
				claims := &auth.Claims{
					UserID: models.NewID(),
					Email:  "test@example.com",
					Role:   models.RoleUser,
				}
				req = requestWithClaims(req, claims)
			}

			rec := httptest.NewRecorder()
			handler.Create(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if !bytes.Contains(rec.Body.Bytes(), []byte(tt.expectedError)) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, rec.Body.String())
			}
		})
	}
}

func TestList_Success_User(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	otherUser := createTestUser(t, testStore, models.RoleUser)

	// Create todos for both users
	todo1 := createTestTodo(t, testStore, user.ID)
	todo2 := createTestTodo(t, testStore, user.ID)
	createTestTodo(t, testStore, otherUser.ID) // should not appear

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req = requestWithClaims(req, claims)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp TodosResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(resp.Todos))
	}

	// Verify we only got our own todos
	for _, todo := range resp.Todos {
		if todo.ID != todo1.ID && todo.ID != todo2.ID {
			t.Errorf("Unexpected todo ID: %s", todo.ID)
		}
	}
}

func TestList_Success_Admin(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	admin := createTestUser(t, testStore, models.RoleAdmin)

	// Create todos for both
	createTestTodo(t, testStore, user.ID)
	createTestTodo(t, testStore, admin.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req = requestWithClaims(req, claims)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp TodosResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Todos) != 2 {
		t.Errorf("Expected 2 todos (all), got %d", len(resp.Todos))
	}
}

func TestList_Unauthorized(t *testing.T) {
	handler, _ := setupTodoTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGetById_Success(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	todo := createTestTodo(t, testStore, user.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.GetById(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp TodoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Todo.ID != todo.ID {
		t.Errorf("Expected todo ID %s, got %s", todo.ID, resp.Todo.ID)
	}
}

func TestGetById_AdminCanViewAny(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	admin := createTestUser(t, testStore, models.RoleAdmin)
	todo := createTestTodo(t, testStore, user.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.GetById(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetById_Forbidden(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	otherUser := createTestUser(t, testStore, models.RoleUser)
	todo := createTestTodo(t, testStore, otherUser.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.GetById(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestGetById_NotFound(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/nonexistent-id", nil)
	req = requestWithClaimsAndID(req, claims, "id", "nonexistent-id")
	rec := httptest.NewRecorder()

	handler.GetById(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	todo := createTestTodo(t, testStore, user.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	newTitle := "Updated Title"
	completed := true
	reqBody := UpdateTodoRequest{
		Title:     &newTitle,
		Completed: &completed,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/todos/"+todo.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp TodoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Todo.Title != newTitle {
		t.Errorf("Expected title %s, got %s", newTitle, resp.Todo.Title)
	}

	if !resp.Todo.Completed {
		t.Error("Expected completed to be true")
	}

	// Description should be unchanged
	if resp.Todo.Description != todo.Description {
		t.Errorf("Expected description %s, got %s", todo.Description, resp.Todo.Description)
	}
}

func TestUpdate_PartialUpdate(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	todo := createTestTodo(t, testStore, user.ID)
	originalTitle := todo.Title

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	// Only update completed
	completed := true
	reqBody := UpdateTodoRequest{
		Completed: &completed,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/todos/"+todo.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp TodoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Todo.Title != originalTitle {
		t.Errorf("Title should be unchanged, expected %s, got %s", originalTitle, resp.Todo.Title)
	}

	if !resp.Todo.Completed {
		t.Error("Expected completed to be true")
	}
}

func TestUpdate_Forbidden(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	otherUser := createTestUser(t, testStore, models.RoleUser)
	todo := createTestTodo(t, testStore, otherUser.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	newTitle := "Hacked"
	reqBody := UpdateTodoRequest{Title: &newTitle}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/todos/"+todo.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestUpdate_AdminCanUpdateAny(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	admin := createTestUser(t, testStore, models.RoleAdmin)
	todo := createTestTodo(t, testStore, user.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   models.RoleAdmin,
	}

	newTitle := "Admin Updated"
	reqBody := UpdateTodoRequest{Title: &newTitle}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/todos/"+todo.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	newTitle := "Whatever"
	reqBody := UpdateTodoRequest{Title: &newTitle}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/todos/nonexistent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDelete_Success_Admin(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	admin := createTestUser(t, testStore, models.RoleAdmin)
	todo := createTestTodo(t, testStore, user.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	// Verify todo is actually deleted
	_, err := testStore.GetTodoByID(todo.ID)
	if err != store.ErrTodoNotFound {
		t.Errorf("Expected todo to be deleted, got error: %v", err)
	}
}

func TestDelete_Forbidden_User(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	user := createTestUser(t, testStore, models.RoleUser)
	todo := createTestTodo(t, testStore, user.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   models.RoleUser,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	// Verify todo still exists
	_, err := testStore.GetTodoByID(todo.ID)
	if err != nil {
		t.Errorf("Todo should still exist, got error: %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	handler, testStore := setupTodoTestHandler(t)

	admin := createTestUser(t, testStore, models.RoleAdmin)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/nonexistent", nil)
	req = requestWithClaimsAndID(req, claims, "id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDelete_Unauthorized(t *testing.T) {
	handler, _ := setupTodoTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/some-id", nil)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
