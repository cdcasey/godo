package store

import "godo/internal/models"

type UserStore interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
}

type TodoStore interface {
	CreateTodo(todo *models.Todo) error
	GetTodoByID(id string) (*models.Todo, error)
	GetTodosByUserID(userID string) ([]*models.Todo, error)
	GetAllTodos() ([]*models.Todo, error)
	UpdateTodo(todo *models.Todo) error
	DeleteTodo(id string) error
}
