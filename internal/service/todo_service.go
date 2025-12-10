package service

import (
	"errors"
	"godo/internal/domain"
	"time"
)

var ErrForbidden = errors.New("forbidden")

type TodoService struct {
	repo domain.TodoRepository
}

func NewTodoService(repo domain.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) Create(userID, title, description string) (*domain.Todo, error) {
	todo := domain.NewTodo(userID, title, description)
	if err := s.repo.Create(todo); err != nil {
		return nil, err
	}
	return todo, nil
}

func (s *TodoService) GetByID(todoID, requestingUserID, requestingUserRole string) (*domain.Todo, error) {
	todo, err := s.repo.GetByID(todoID)
	if err != nil {
		return nil, err
	}

	if requestingUserRole != domain.RoleAdmin && todo.UserID != requestingUserID {
		return nil, ErrForbidden
	}

	return todo, nil
}

func (s *TodoService) List(requestingUserID, requestingUserRole string) ([]*domain.Todo, error) {
	if requestingUserRole == domain.RoleAdmin {
		return s.repo.GetAll()
	}

	return s.repo.GetByUserID(requestingUserID)
}

func (s *TodoService) Update(todoID, requestingUserID, requestingUserRole string, title, description *string, completed *bool) (*domain.Todo, error) {
	todo, err := s.repo.GetByID(todoID)
	if err != nil {
		return nil, err
	}

	if requestingUserRole != domain.RoleAdmin && todo.UserID != requestingUserID {
		return nil, ErrForbidden
	}

	if title != nil {
		todo.Title = *title
	}
	if description != nil {
		todo.Description = *description
	}
	if completed != nil {
		todo.Completed = *completed
	}
	todo.UpdatedAt = time.Now()

	if err := s.repo.Update(todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *TodoService) Delete(todoID, requestingUserID, requestingUserRole string) error {
	if requestingUserRole != domain.RoleAdmin {
		return ErrForbidden
	}

	return s.repo.Delete(todoID)
}
