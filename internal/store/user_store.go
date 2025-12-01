package store

import (
	"database/sql"
	"fmt"
	"godo/internal/models"
)

func (s *Store) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, email, password_hash, role, created_at)
						VALUES (?,?,?,?,?)
	`

	_, err := s.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, created_at
						FROM users WHERE email = ?`

	var user models.User
	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (s *Store) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, created_at
						FROM users where id = ?`

	var user models.User
	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}
