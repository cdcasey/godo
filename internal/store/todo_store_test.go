package store

import (
	"godo/internal/models"
	"testing"
)

func TestCreateTodo_Success(t *testing.T) {
	store := setupTestStore(t)

	// Create a user first (foreign key constraint)
	user := &models.User{
		ID:           models.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleUser,
	}
	if err := store.CreateUser(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	todo := models.NewTodo(user.ID, "Test title", "Test description")

	err := store.CreateTodo(todo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was saved
	saved, err := store.GetTodoByID(todo.ID)
	if err != nil {
		t.Fatalf("failed to retrieve todo: %v", err)
	}

	if saved.Title != todo.Title {
		t.Errorf("expected title %q, got %q", todo.Title, saved.Title)
	}
}

func TestGetTodoByID_Success(t *testing.T) {
	store := setupTestStore(t)

	user := &models.User{
		ID:           models.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleUser,
	}
	store.CreateUser(user)

	todo := models.NewTodo(user.ID, "Test title", "Test description")
	store.CreateTodo(todo)

	found, err := store.GetTodoByID(todo.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if found.ID != todo.ID {
		t.Errorf("expected ID %q, got %q", todo.ID, found.ID)
	}
	if found.UserID != user.ID {
		t.Errorf("expected UserID %q, got %q", user.ID, found.UserID)
	}
}

func TestGetTodoByID_NotFound(t *testing.T) {
	store := setupTestStore(t)

	_, err := store.GetTodoByID("nonexistent-id")
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestGetTodosByUserID_Success(t *testing.T) {
	store := setupTestStore(t)

	user := &models.User{
		ID:           models.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleUser,
	}
	store.CreateUser(user)

	todo1 := models.NewTodo(user.ID, "First", "")
	todo2 := models.NewTodo(user.ID, "Second", "")
	store.CreateTodo(todo1)
	store.CreateTodo(todo2)

	todos, err := store.GetTodosByUserID(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(todos) != 2 {
		t.Errorf("expected 2 todos, got %d", len(todos))
	}
}

func TestGetTodosByUserID_Empty(t *testing.T) {
	store := setupTestStore(t)

	user := &models.User{
		ID:           models.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleUser,
	}
	store.CreateUser(user)

	todos, err := store.GetTodosByUserID(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}
}

func TestUpdateTodo_Success(t *testing.T) {
	store := setupTestStore(t)

	user := &models.User{
		ID:           models.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleUser,
	}
	store.CreateUser(user)

	todo := models.NewTodo(user.ID, "Original", "Original desc")
	store.CreateTodo(todo)

	todo.Title = "Updated"
	todo.Description = "Updated desc"
	todo.Completed = true

	err := store.UpdateTodo(todo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := store.GetTodoByID(todo.ID)
	if updated.Title != "Updated" {
		t.Errorf("expected title %q, got %q", "Updated", updated.Title)
	}
	if updated.Completed != true {
		t.Error("expected completed to be true")
	}
}

func TestUpdateTodo_NotFound(t *testing.T) {
	store := setupTestStore(t)

	todo := &models.Todo{ID: "nonexistent-id"}

	err := store.UpdateTodo(todo)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestDeleteTodo_Success(t *testing.T) {
	store := setupTestStore(t)

	user := &models.User{
		ID:           models.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleUser,
	}
	store.CreateUser(user)

	todo := models.NewTodo(user.ID, "To delete", "")
	store.CreateTodo(todo)

	err := store.DeleteTodo(todo.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = store.GetTodoByID(todo.ID)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound after delete, got %v", err)
	}
}

func TestDeleteTodo_NotFound(t *testing.T) {
	store := setupTestStore(t)

	err := store.DeleteTodo("nonexistent-id")
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}
