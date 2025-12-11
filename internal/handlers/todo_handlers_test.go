package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"godo/internal/auth"
	"godo/internal/domain"
	"godo/internal/service"
	"godo/internal/store"
	"godo/internal/testutil"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
)

func setupTodoTestHandler(t *testing.T) (*TodoHandler, *store.UserRepo, *store.TodoRepo) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	userRepo := store.NewUserRepo(db)
	todoRepo := store.NewTodoRepo(db)
	todoService := service.NewTodoService(todoRepo)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := NewTodoHandler(todoService, logger)

	return handler, userRepo, todoRepo
}

func createTestUser(t *testing.T, userRepo *store.UserRepo, role string) *domain.User {
	t.Helper()
	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test-" + domain.NewID() + "@example.com",
		PasswordHash: "hashed",
		Role:         role,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createTestTodo(t *testing.T, todoRepo *store.TodoRepo, userID string) *domain.Todo {
	t.Helper()
	todo := domain.NewTodo(userID, "Test Todo", "Test Description")
	if err := todoRepo.Create(todo); err != nil {
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
	handler, userRepo, _ := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	reqBody := CreateTodoRequest{
		Title:       "My Todo",
		Description: "My Description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaims(req, claims)
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

	if resp.Todo.UserID != user.ID {
		t.Errorf("Expected user_id %s, got %s", user.ID, resp.Todo.UserID)
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
			handler, _, _ := setupTodoTestHandler(t)

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
					UserID: domain.NewID(),
					Email:  "test@example.com",
					Role:   domain.RoleUser,
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
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	otherUser := createTestUser(t, userRepo, domain.RoleUser)

	todo1 := createTestTodo(t, todoRepo, user.ID)
	todo2 := createTestTodo(t, todoRepo, user.ID)
	createTestTodo(t, todoRepo, otherUser.ID) // should not appear

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
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

	for _, todo := range resp.Todos {
		if todo.ID != todo1.ID && todo.ID != todo2.ID {
			t.Errorf("Unexpected todo ID: %s", todo.ID)
		}
	}
}

func TestList_Success_Admin(t *testing.T) {
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	admin := createTestUser(t, userRepo, domain.RoleAdmin)

	createTestTodo(t, todoRepo, user.ID)
	createTestTodo(t, todoRepo, admin.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
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
	handler, _, _ := setupTodoTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGetByID_Success(t *testing.T) {
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	todo := createTestTodo(t, todoRepo, user.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

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

func TestGetByID_AdminCanViewAny(t *testing.T) {
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	todo := createTestTodo(t, todoRepo, user.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetByID_Forbidden(t *testing.T) {
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	otherUser := createTestUser(t, userRepo, domain.RoleUser)
	todo := createTestTodo(t, todoRepo, otherUser.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	handler, userRepo, _ := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos/nonexistent-id", nil)
	req = requestWithClaimsAndID(req, claims, "id", "nonexistent-id")
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	todo := createTestTodo(t, todoRepo, user.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
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

	if resp.Todo.Description != todo.Description {
		t.Errorf("Expected description %s, got %s", todo.Description, resp.Todo.Description)
	}
}

func TestUpdate_PartialUpdate(t *testing.T) {
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	todo := createTestTodo(t, todoRepo, user.ID)
	originalTitle := todo.Title

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

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
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	otherUser := createTestUser(t, userRepo, domain.RoleUser)
	todo := createTestTodo(t, todoRepo, otherUser.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
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
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	todo := createTestTodo(t, todoRepo, user.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
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
	handler, userRepo, _ := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
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
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	todo := createTestTodo(t, todoRepo, user.ID)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	_, err := todoRepo.GetByID(todo.ID)
	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected todo to be deleted, got error: %v", err)
	}
}

func TestDelete_Forbidden_User(t *testing.T) {
	handler, userRepo, todoRepo := setupTodoTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	todo := createTestTodo(t, todoRepo, user.ID)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/"+todo.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", todo.ID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	_, err := todoRepo.GetByID(todo.ID)
	if err != nil {
		t.Errorf("Todo should still exist, got error: %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	handler, userRepo, _ := setupTodoTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
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
	handler, _, _ := setupTodoTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/todos/some-id", nil)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
