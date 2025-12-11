package store

import (
	"database/sql"
	"fmt"
	"godo/internal/domain"
)

// UserRepo Note to self: this implements the UserRepository interface by having all of the required methods.
// Go does not have an explicit "implements" keyword.
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *domain.User) error {
	query := `INSERT INTO users (id, email, password_hash, role, created_at)
		VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, role, created_at
		FROM users WHERE email = ?`

	var user domain.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetByID(id string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, role, created_at
			  FROM users WHERE id = ?`

	var user domain.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetAll() ([]*domain.User, error) {
	query := `SELECT id, email, role, created_at
		FROM users ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan users: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepo) Update(user *domain.User) error {
	query := `UPDATE users SET email = ?, password_hash = ?, role = ? WHERE id = ?`

	result, err := r.db.Exec(query, user.Email, user.PasswordHash, user.Role, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) Delete(id string) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) CountByRole(role string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE role = ?`

	var userCount int
	err := r.db.QueryRow(query, role).Scan(
		&userCount,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to scan users: %w", err)
	}

	return userCount, nil
}
