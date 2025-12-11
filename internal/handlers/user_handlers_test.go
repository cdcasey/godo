package handlers

import (
	"bytes"
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
)

func setupUserTestHandler(t *testing.T) (*UserHandler, *store.UserRepo) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	userRepo := store.NewUserRepo(db)
	userService := service.NewUserService(userRepo)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := NewUserHandler(userService, logger)

	return handler, userRepo
}

// TestUserList_Success_Admin verifies admins can list all users
func TestUserList_Success_Admin(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	user1 := createTestUser(t, userRepo, domain.RoleUser)
	user2 := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req = requestWithClaims(req, claims)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp UsersResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(resp.Users))
	}

	// Verify all users are present
	ids := make(map[string]bool)
	for _, u := range resp.Users {
		ids[u.ID] = true
	}
	for _, id := range []string{admin.ID, user1.ID, user2.ID} {
		if !ids[id] {
			t.Errorf("Expected user %s in response", id)
		}
	}
}

// TestList_Forbidden_User verifies regular users cannot list users
func TestUserList_Forbidden_User(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req = requestWithClaims(req, claims)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

// TestUserList_Unauthorized verifies unauthenticated requests are rejected
func TestUserList_Unauthorized(t *testing.T) {
	handler, _ := setupUserTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

// TestGetByID_Success_Self verifies users can view their own profile
func TestUserGetByID_Success_Self(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+user.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.User.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, resp.User.ID)
	}
}

// TestGetByID_Success_Admin verifies admins can view any user
func TestUserGetByID_Success_Admin(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+user.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.User.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, resp.User.ID)
	}
}

// TestGetByID_Forbidden verifies users cannot view other users
func TestUserGetByID_Forbidden(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	otherUser := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+otherUser.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", otherUser.ID)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

// TestGetByID_NotFound verifies 404 for non-existent users
func TestUserGetByID_NotFound(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/nonexistent", nil)
	req = requestWithClaimsAndID(req, claims, "id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// TestGetByID_Unauthorized verifies unauthenticated requests are rejected
func TestUserGetByID_Unauthorized(t *testing.T) {
	handler, _ := setupUserTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/some-id", nil)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

// TestUpdate_Success_Self verifies users can update their own profile
func TestUserUpdate_Success_Self(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	newEmail := "updated@example.com"
	reqBody := UpdateUserRequest{
		Email: &newEmail,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/"+user.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.User.Email != newEmail {
		t.Errorf("Expected email %s, got %s", newEmail, resp.User.Email)
	}
}

// TestUpdate_Success_Admin verifies admins can update any user
func TestUserUpdate_Success_Admin(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	newEmail := "admin-updated@example.com"
	reqBody := UpdateUserRequest{
		Email: &newEmail,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/"+user.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.User.Email != newEmail {
		t.Errorf("Expected email %s, got %s", newEmail, resp.User.Email)
	}
}

// TestUpdate_AdminCanChangeRole verifies admins can change user roles
func TestUserUpdate_AdminCanChangeRole(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	newRole := domain.RoleAdmin
	reqBody := UpdateUserRequest{
		Role: &newRole,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/"+user.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.User.Role != domain.RoleAdmin {
		t.Errorf("Expected role %s, got %s", domain.RoleAdmin, resp.User.Role)
	}
}

// TestUpdate_Forbidden_OtherUser verifies users cannot update others
func TestUserUpdate_Forbidden_OtherUser(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	otherUser := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	newEmail := "hacked@example.com"
	reqBody := UpdateUserRequest{
		Email: &newEmail,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/"+otherUser.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", otherUser.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

// TestUpdate_UserCannotChangeOwnRole verifies users cannot escalate privileges
func TestUserUpdate_UserCannotChangeOwnRole(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	newRole := domain.RoleAdmin
	reqBody := UpdateUserRequest{
		Role: &newRole,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/"+user.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

// TestUpdate_CannotDemoteLastAdmin verifies the last admin cannot be demoted
func TestUserUpdate_CannotDemoteLastAdmin(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	// Create only one admin
	admin := createTestUser(t, userRepo, domain.RoleAdmin)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	newRole := domain.RoleUser
	reqBody := UpdateUserRequest{
		Role: &newRole,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/"+admin.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", admin.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte("Cannot demote the last admin")) {
		t.Errorf("Expected error about last admin, got %s", rec.Body.String())
	}
}

// TestUserUpdate_NotFound verifies 404 for non-existent users
func TestUserUpdate_NotFound(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	newEmail := "whatever@example.com"
	reqBody := UpdateUserRequest{
		Email: &newEmail,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/nonexistent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// TestUpdate_InvalidJSON verifies bad request for invalid JSON
func TestUserUpdate_InvalidJSON(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/users/"+user.ID, bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestUpdate_Unauthorized verifies unauthenticated requests are rejected
func TestUserUpdate_Unauthorized(t *testing.T) {
	handler, _ := setupUserTestHandler(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/users/some-id", nil)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

// TestUserDelete_Success_Admin verifies admins can delete users
func TestUserDelete_Success_Admin(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)
	user := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/"+user.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", user.ID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	// Verify user is actually deleted
	_, err := userRepo.GetByID(user.ID)
	if err != store.ErrUserNotFound {
		t.Errorf("Expected user to be deleted, got error: %v", err)
	}
}

// TestUserDelete_Forbidden_User verifies regular users cannot delete users
func TestUserDelete_Forbidden_User(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	user := createTestUser(t, userRepo, domain.RoleUser)
	otherUser := createTestUser(t, userRepo, domain.RoleUser)

	claims := &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   domain.RoleUser,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/"+otherUser.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", otherUser.ID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	// Verify user still exists
	_, err := userRepo.GetByID(otherUser.ID)
	if err != nil {
		t.Errorf("User should still exist, got error: %v", err)
	}
}

// TestDelete_CannotDeleteLastAdmin verifies the last admin cannot be deleted
func TestUserDelete_CannotDeleteLastAdmin(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	// Create only one admin
	admin := createTestUser(t, userRepo, domain.RoleAdmin)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/"+admin.ID, nil)
	req = requestWithClaimsAndID(req, claims, "id", admin.ID)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte("Cannot delete the last admin")) {
		t.Errorf("Expected error about last admin, got %s", rec.Body.String())
	}
}

// TestUserDelete_NotFound verifies 404 for non-existent users
func TestUserDelete_NotFound(t *testing.T) {
	handler, userRepo := setupUserTestHandler(t)

	admin := createTestUser(t, userRepo, domain.RoleAdmin)

	claims := &auth.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   domain.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/nonexistent", nil)
	req = requestWithClaimsAndID(req, claims, "id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// TestUserDelete_Unauthorized verifies unauthenticated requests are rejected
func TestUserDelete_Unauthorized(t *testing.T) {
	handler, _ := setupUserTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/some-id", nil)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
