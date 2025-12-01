package store

import (
	"godo/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateUser_Success(t *testing.T) {
	store := setupTestStore(t)

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
		CreatedAt:    time.Now(),
	}

	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("Error creating user: %v", err)
	}
}

func TestGetUserByEmail_Success(t *testing.T) {
	store := setupTestStore(t)

	user := &models.User{
		ID:           uuid.NewString(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
		CreatedAt:    time.Now(),
	}

	if err := store.CreateUser(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := store.GetUserByEmail("test@example.com")
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

func TestGetUserByEmail_NotFound(t *testing.T) {
	store := setupTestStore(t)

	user, err := store.GetUserByEmail("nonexistent@exaple.com")

	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}
