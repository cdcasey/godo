package store

import "godo/internal/models"

type UserStore interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
}

type ToDoStore interface {
	CreateTodo(todo *models.ToDo) error
	GetToDoByID(id string) (*models.ToDo, error)
	GetToDosByUserID(userID string) ([]*models.ToDo, error)
	UpdateToDo(todo *models.ToDo) error
	DeleteToDO(id string) error
}
