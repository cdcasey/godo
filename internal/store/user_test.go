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
