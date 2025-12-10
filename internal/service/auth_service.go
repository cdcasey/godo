package service

import (
	"errors"
	"godo/internal/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
)

type AuthService struct {
	repo domain.UserRepository
}

func NewAuthService(repo domain.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(email, password string) (*domain.User, error) {
	hashedPassword, err := domain.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           domain.NewId(),
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         domain.RoleUser,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Authenticate(email, password string) (*domain.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
