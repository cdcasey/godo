package store

import (
	"database/sql"
	"fmt"
	"godo/internal/models"
)

func (s *Store) CreateTodo(todo *models.Todo) error {
	query := `INSERT INTO todos (id, user_id, title, description, completed, created_at, updated_at)
						VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, todo.ID, todo.UserID, todo.Title, todo.Description, todo.Completed, todo.CreatedAt, todo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	return nil
}

func (s *Store) GetTodoByID(id string) (*models.Todo, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
            FROM todos WHERE id = ?`

	var todo models.Todo
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
		return nil, ErrTodoNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return &todo, nil
}

func (s *Store) GetTodosByUserID(userID string) ([]*models.Todo, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
						FROM todos WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)

	}
	defer rows.Close()

	var todos []*models.Todo
	for rows.Next() {
		var todo models.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, &todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating todos: %w", err)
	}

	return todos, nil
}

func (s *Store) UpdateTodo(todo *models.Todo) error {
	query := `UPDATE todos SET title = ?, description = ?, completed = ?, updated_at = ?
					WHERE id = ?`

	result, err := s.db.Exec(query, todo.Title, todo.Description, todo.Completed, todo.UpdatedAt, todo.ID)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrTodoNotFound
	}

	return nil
}

func (s *Store) DeleteTodo(id string) error {
	query := `DELETE FROM todos WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrTodoNotFound
	}

	return nil
}
