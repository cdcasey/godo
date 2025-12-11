package service

import "godo/internal/domain"

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
