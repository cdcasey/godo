package store

import (
	"database/sql"
	"fmt"
	"godo/internal/models"
)

func (s *Store) CreateTodo(todo *models.ToDo) error {
	query := `INSERT INTO todos (id, user_id, title, description, completed, created_at, updated_at)
						VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, todo.ID, todo.UserID, todo.Title, todo.Description, todo.Completed, todo.CreatedAt, todo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	return nil
}

func (s *Store) GetToDoByID(id string) (*models.ToDo, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
            FROM todos WHERE id = ?`

	var todo models.ToDo
	err := s.db.QueryRow(query, id).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrToDoNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return &todo, nil
}
