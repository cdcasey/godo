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
func TestUserServiceUpdate_User_Success(t *testing.T) {
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

	newEmail := "test2@example.com"

	user, err := userService.Update(user.ID, user.ID, user.Role, &newEmail, nil, nil)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if user.Email != newEmail {
		t.Fatalf("Expected email %s got %s", newEmail, user.Email)
	}
}

func TestUserServiceUpdate_User_Failure(t *testing.T) {
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

	newEmail := "test2@example.com"

	user, err := userService.Update(user.ID, domain.NewID(), domain.RoleUser, &newEmail, nil, nil)
	if err != ErrForbidden {
		t.Fatalf("Expected ErrForbidden got: %v", err)
	}

	if user != nil {
		t.Fatalf("Expected nil user, got: %v", user)
	}
}

func TestUserServiceUpdate_UserRole_Failure(t *testing.T) {
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

	newRole := domain.RoleAdmin

	user, err := userService.Update(user.ID, user.ID, domain.RoleUser, nil, nil, &newRole)
	if err != ErrForbidden {
		t.Fatalf("Expected ErrForbidden got: %v", err)
	}
}

func TestUserServiceUpdate_Admin_Success(t *testing.T) {
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

	newRole := domain.RoleAdmin

	updatedUser, err := userService.Update(user.ID, domain.NewID(), domain.RoleAdmin, nil, nil, &newRole)
	if err != nil {
		t.Fatalf("Expected to update user got: %v", err)
	}

	if updatedUser.Role != newRole {
		t.Fatalf("Expected role %s got: %s", newRole, updatedUser.Role)
	}
}

func TestUserServiceUpdate_AdminDemote_Success(t *testing.T) {
	userService, userRepo := setupTestUserService(t)

	user1 := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleAdmin,
	}
	user2 := &domain.User{
		ID:           domain.NewID(),
		Email:        "test2@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleAdmin,
	}

	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	newRole := domain.RoleUser

	updatedUser, err := userService.Update(user1.ID, domain.NewID(), domain.RoleAdmin, nil, nil, &newRole)
	if err != nil {
		t.Fatalf("Expected to update user got: %v", err)
	}

	if updatedUser.Role != newRole {
		t.Fatalf("Expected role %s got: %s", newRole, updatedUser.Role)
	}
}

func TestUserServiceUpdate_AdminDemote_Failure(t *testing.T) {
	userService, userRepo := setupTestUserService(t)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleAdmin,
	}

	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	newRole := domain.RoleUser

	user, err := userService.Update(user.ID, user.ID, user.Role, nil, nil, &newRole)
	if err != ErrLastAdmin {
		t.Fatalf("Expected ErrLastAdmin got: %v", err)
	}
}

// Delete: user can delete self, admin can delete anyone, can't delete last admin
func TestUserServiceDelete_User_Success(t *testing.T) {
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

	err := userService.Delete(user.ID, user.ID, user.Role)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
}

func TestUserServiceDelete_User_Failure(t *testing.T) {
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

	err := userService.Delete(user.ID, domain.NewID(), user.Role)
	if err != ErrForbidden {
		t.Fatalf("Expected ErrForbidden, got: %v", err)
	}
}

func TestUserServiceDelete_Admin_Success(t *testing.T) {
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

	err := userService.Delete(user.ID, domain.NewID(), domain.RoleAdmin)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
}

func TestUserServiceDelete_Admin_Failure(t *testing.T) {
	userService, userRepo := setupTestUserService(t)

	user := &domain.User{
		ID:           domain.NewID(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         domain.RoleAdmin,
	}

	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err := userService.Delete(user.ID, user.ID, user.Role)
	if err != ErrLastAdmin {
		t.Fatalf("Expected ErrLastAdmin, got: %v", err)
	}
}
