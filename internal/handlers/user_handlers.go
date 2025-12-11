package handlers

import (
	"encoding/json"
	"errors"
	"godo/internal/auth"
	"godo/internal/domain"
	"godo/internal/service"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userService *service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	Role     *string `json:"role,omitempty"`
}

type UserResponse struct {
	User domain.User `json:"user"`
}

type UsersResponse struct {
	Users []*domain.User `json:"users"`
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := h.userService.List(claims.Role)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to list users", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Users listed", "user_id", claims.UserID, "count", len(users))

	writeJsonResponse(w, http.StatusOK, UsersResponse{Users: users}, h.logger)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(userID, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to get user", "error", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, http.StatusOK, UserResponse{User: *user}, h.logger)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Update(userID, claims.UserID, claims.Role, req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrLastAdmin) {
			http.Error(w, "Cannot demote the last admin", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to update user", "error", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User updated", "user_id", userID, "requesting_user_id", claims.UserID)

	writeJsonResponse(w, http.StatusOK, UserResponse{User: *user}, h.logger)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	err := h.userService.Delete(userID, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrLastAdmin) {
			http.Error(w, "Cannot delete the last admin", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to delete user", "error", err, "user_id", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User deleted", "user_id", userID, "requesting_user_id", claims.UserID)

	w.WriteHeader(http.StatusNoContent)
}
