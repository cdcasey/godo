package store

import (
	"godo/internal/domain"
	"testing"
)

func TestUserRepo_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Error creating user: %v", err)
	}
}

func TestUserRepo_GetByEmail_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
	}

	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := repo.GetByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected user, got nil")
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
	}

	if retrieved.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, retrieved.Role)
	}
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	user, err := repo.GetByEmail("nonexistent@example.com")

	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}

func TestUserRepo_GetByID_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
	}

	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected user, got nil")
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}

	if retrieved.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, retrieved.Role)
	}
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	user, err := repo.GetByID(domain.NewID())

	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}

func TestUserRepo_Update_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	updatedEmail := "test2@example.com"
	updatedPasswordHash := "hash2"
	updatedRole := domain.RoleAdmin

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user)

	user.Email = updatedEmail
	user.PasswordHash = updatedPasswordHash
	user.Role = updatedRole

	err := userRepo.Update(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, err := userRepo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("failed to retrieve updated user: %v", err)
	}

	if updated.Email != updatedEmail {
		t.Errorf("expected email %q, got %q", updatedEmail, updated.Email)
	}

	if updated.PasswordHash != updatedPasswordHash {
		t.Errorf("expected hash %q, got %q", updatedPasswordHash, updated.PasswordHash)
	}

	if updated.Role != updatedRole {
		t.Errorf("expected role %q, got %q", updatedRole, updated.Role)
	}
}

func TestUserRepo_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)

	user := &domain.User{ID: "nonexistent-id"}

	err := userRepo.Update(user)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepo_Delete_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user)

	err := userRepo.Delete(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = userRepo.GetByID(user.ID)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound after delete, got %v", err)
	}
}

func TestUserRepo_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)

	err := userRepo.Delete("nonexistent-id")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
