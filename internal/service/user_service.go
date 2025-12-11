package service

import (
	"godo/internal/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetById(userID, requestingUserID, requestingUserRole string) (*domain.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if requestingUserRole != domain.RoleAdmin && user.ID != requestingUserID {
		return nil, ErrForbidden
	}

	return user, nil
}

func (s *UserService) List(requestingUserRole string) ([]*domain.User, error) {
	if requestingUserRole == domain.RoleAdmin {
		return s.repo.GetAll()
	}

	return nil, ErrForbidden
}

func (s *UserService) Update(userID, requestingUserID, requestingUserRole string, newEmail, newPassword, newRole *string) (*domain.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if requestingUserRole != domain.RoleAdmin && user.ID != requestingUserID {
		return nil, ErrForbidden
	}

	if newEmail != nil {
		user.Email = *newEmail
	}
	if newPassword != nil {
		hashedPassword, err := domain.HashPassword(*newPassword)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}
	if newRole != nil {
		if requestingUserRole == domain.RoleUser {
			return nil, ErrForbidden
		}
		if user.Role == domain.RoleAdmin && *newRole == domain.RoleUser {
			count, err := s.repo.CountByRole(domain.RoleAdmin)
			if err != nil {
				return nil, err
			}
			if count < 2 {
				return nil, ErrLastAdmin
			}
		}
		user.Role = *newRole
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(userID, requestingUserID, requestingUserRole string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	if requestingUserRole != domain.RoleAdmin && user.ID != requestingUserID {
		return ErrForbidden
	}

	if user.Role == domain.RoleAdmin {
		count, err := s.repo.CountByRole(domain.RoleAdmin)
		if err != nil {
			return err
		}
		if count == 1 {
			return ErrLastAdmin
		}
	}

	return s.repo.Delete(userID)
}
