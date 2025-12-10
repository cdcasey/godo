package store

import (
	"godo/internal/domain"
	"testing"
)

func TestTodoRepo_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	todoRepo := NewTodoRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	todo := domain.NewTodo(user.ID, "Test title", "Test description")

	err := todoRepo.Create(todo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	saved, err := todoRepo.GetByID(todo.ID)
	if err != nil {
		t.Fatalf("failed to retrieve todo: %v", err)
	}

	if saved.Title != todo.Title {
		t.Errorf("expected title %q, got %q", todo.Title, saved.Title)
	}
}

func TestTodoRepo_GetByID_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	todoRepo := NewTodoRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user)

	todo := domain.NewTodo(user.ID, "Test title", "Test description")
	todoRepo.Create(todo)

	found, err := todoRepo.GetByID(todo.ID)
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

func TestTodoRepo_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	todoRepo := NewTodoRepo(db)

	_, err := todoRepo.GetByID("nonexistent-id")
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestTodoRepo_GetByUserID_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	todoRepo := NewTodoRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user)

	todo1 := domain.NewTodo(user.ID, "First", "")
	todo2 := domain.NewTodo(user.ID, "Second", "")
	todoRepo.Create(todo1)
	todoRepo.Create(todo2)

	todos, err := todoRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(todos) != 2 {
		t.Errorf("expected 2 todos, got %d", len(todos))
	}
}

func TestTodoRepo_GetByUserID_Empty(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	todoRepo := NewTodoRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user)

	todos, err := todoRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}
}

func TestTodoRepo_GetAll_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	todoRepo := NewTodoRepo(db)

	user1 := &domain.User{
		ID:           domain.NewID(),
		Email:        "user1@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	user2 := &domain.User{
		ID:           domain.NewID(),
		Email:        "user2@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user1)
	userRepo.Create(user2)

	todoRepo.Create(domain.NewTodo(user1.ID, "User1 Todo", ""))
	todoRepo.Create(domain.NewTodo(user2.ID, "User2 Todo", ""))

	todos, err := todoRepo.GetAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(todos) != 2 {
		t.Errorf("expected 2 todos, got %d", len(todos))
	}
}

func TestTodoRepo_Update_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	todoRepo := NewTodoRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user)

	todo := domain.NewTodo(user.ID, "Original", "Original desc")
	todoRepo.Create(todo)

	todo.Title = "Updated"
	todo.Description = "Updated desc"
	todo.Completed = true

	err := todoRepo.Update(todo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := todoRepo.GetByID(todo.ID)
	if updated.Title != "Updated" {
		t.Errorf("expected title %q, got %q", "Updated", updated.Title)
	}
	if updated.Completed != true {
		t.Error("expected completed to be true")
	}
}

func TestTodoRepo_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	todoRepo := NewTodoRepo(db)

	todo := &domain.Todo{ID: "nonexistent-id"}

	err := todoRepo.Update(todo)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestTodoRepo_Delete_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	todoRepo := NewTodoRepo(db)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
	}
	userRepo.Create(user)

	todo := domain.NewTodo(user.ID, "To delete", "")
	todoRepo.Create(todo)

	err := todoRepo.Delete(todo.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = todoRepo.GetByID(todo.ID)
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound after delete, got %v", err)
	}
}

func TestTodoRepo_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	todoRepo := NewTodoRepo(db)

	err := todoRepo.Delete("nonexistent-id")
	if err != ErrTodoNotFound {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}
