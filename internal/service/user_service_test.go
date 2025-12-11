package service

import (
	"godo/internal/domain"
	"godo/internal/store"
	"godo/internal/testutil"
	"testing"
)

func setupTestUserService(t *testing.T) (*UserService, *store.UserRepo) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	userRepo := store.NewUserRepo(db)
	userService := NewUserService(userRepo)

	return userService, userRepo
}

// GetByID: user can get self, admin can get anyone, user can't get others
func TestUserServiceGetById_UserGetsSelf_Success(t *testing.T) {
	userService, userRepo := setupTestUserService(t)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
	}

	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := userService.GetByID(user.ID, user.ID, user.Role)
	if err != nil {
		t.Fatalf("Failed to get user by id: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected user, got nil")
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
	}
}

func TestUserServiceGetById_UserGetsOther_Failure(t *testing.T) {
	userService, userRepo := setupTestUserService(t)

	requestingUserID := domain.NewID()
	requestingUserRole := domain.RoleUser

	targetUser := &domain.User{
		ID:           domain.NewID(),
		Email:        "test2@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
	}

	if err := userRepo.Create(targetUser); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := userService.GetByID(targetUser.ID, requestingUserID, requestingUserRole)
	if err != ErrForbidden {
		t.Fatalf("Expected ErrForbidden got: %v", err)
	}

	if retrieved != nil {
		t.Fatalf("Expected nil, got user: %v", retrieved)
	}
}

func TestUserServiceGetById_AdminGetsAny(t *testing.T) {
	userService, userRepo := setupTestUserService(t)

	requestingUserID := domain.NewID()
	requestingUserRole := domain.RoleAdmin

	targetUser := &domain.User{
		ID:           domain.NewID(),
		Email:        "test2@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
	}

	if err := userRepo.Create(targetUser); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := userService.GetByID(targetUser.ID, requestingUserID, requestingUserRole)
	if err != nil {
		t.Fatalf("Expected user got: %v", err)
	}

	if retrieved.ID != targetUser.ID {
		t.Fatalf("Expected ID %s, got %s", targetUser.ID, retrieved.ID)
	}
}

func TestUserServiceGetById_NotFound(t *testing.T) {
	userService, _ := setupTestUserService(t)

	requestingUserID := domain.NewID()
	requestingUserRole := domain.RoleUser

	retrieved, err := userService.GetByID(domain.NewID(), requestingUserID, requestingUserRole)
	if err != store.ErrUserNotFound {
		t.Fatalf("Expected ErrUserNotFound got: %v", err)
	}

	if retrieved != nil {
		t.Fatalf("Expected nil, got user: %v", retrieved)
	}
}

// List: admin only
func TestUserServiceList_Success(t *testing.T) {
	userService, userRepo := setupTestUserService(t)

	requestingUserRole := domain.RoleAdmin

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test2@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
	}

	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	userList, err := userService.List(requestingUserRole)
	if err != nil {
		t.Fatalf("Expected users got: %v", err)
	}

	if len(userList) != 1 {
		t.Fatalf("Expected a list of one user, got: %v", len(userList))
	}
}

func TestUserServiceList_Failure(t *testing.T) {
	userService, _ := setupTestUserService(t)
	userList, err := userService.List(domain.RoleUser)
	if err != ErrForbidden {
		t.Fatalf("Expected ErrForbidden got: %v", err)
	}
	if userList != nil {
		t.Fatalf("Expected nil, got: %v", userList)
	}
}

// Update: user can update self (but not role), admin can update anyone, can't demote last admin
// Delete: user can delete self, admin can delete anyone, can't delete last admin
