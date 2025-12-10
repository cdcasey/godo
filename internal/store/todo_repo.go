package store

import (
	"database/sql"
	"fmt"
	"godo/internal/domain"
)

type TodoRepo struct {
	db *sql.DB
}

func NewTodoRepo(db *sql.DB) *TodoRepo {
	return &TodoRepo{db: db}
}

func (r *TodoRepo) Create(todo *domain.Todo) error {
	query := `INSERT INTO todos (id, user_id, title, description, completed, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, todo.ID, todo.UserID, todo.Title, todo.Description, todo.Completed, todo.CreatedAt, todo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	return nil
}

func (r *TodoRepo) GetByID(id string) (*domain.Todo, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
			  FROM todos WHERE id = ?`

	var todo domain.Todo
	err := r.db.QueryRow(query, id).Scan(
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

func (r *TodoRepo) GetByUserID(userID string) ([]*domain.Todo, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}
	defer rows.Close()

	todos := make([]*domain.Todo, 0)
	for rows.Next() {
		var todo domain.Todo
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

func (r *TodoRepo) GetAll() ([]*domain.Todo, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
			  FROM todos ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}
	defer rows.Close()

	todos := make([]*domain.Todo, 0)
	for rows.Next() {
		var todo domain.Todo
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

func (r *TodoRepo) Update(todo *domain.Todo) error {
	query := `UPDATE todos SET title = ?, description = ?, completed = ?, updated_at = ?
			  WHERE id = ?`

	result, err := r.db.Exec(query, todo.Title, todo.Description, todo.Completed, todo.UpdatedAt, todo.ID)
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

func (r *TodoRepo) Delete(id string) error {
	query := `DELETE FROM todos WHERE id = ?`

	result, err := r.db.Exec(query, id)
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
