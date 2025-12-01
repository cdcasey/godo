package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // - means never serialize password hash
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)
