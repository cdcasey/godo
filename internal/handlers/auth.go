package handlers

import (
	"encoding/json"
	"errors"
	"godo/internal/auth"
	"godo/internal/models"
	"godo/internal/store"
	"log/slog"
	"net/http"
	"time"
)

type AuthHandler struct {
	store     *store.Store
	logger    *slog.Logger
	jwtSecret string
}

func NewAuthHandler(store *store.Store, logger *slog.Logger, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		store:     store,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Failed to hash password", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		ID:           models.NewID(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         models.RoleUser,
		CreatedAt:    time.Now(),
	}

	if err := h.store.CreateUser(user); err != nil {
		h.logger.Error("Failed to create user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User registered", "user_id", user.ID)

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, h.jwtSecret, 24*time.Hour)
	if err != nil {
		h.logger.Error("Failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			h.logger.Warn("Login attempt with non-existent email", "email", req.Email)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		h.logger.Error("Failed to get user", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !user.CheckPassword(req.Password) {
		h.logger.Warn("Login attempt with incorrect password", "email", req.Email)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, h.jwtSecret, 24*time.Hour)
	if err != nil {
		h.logger.Error("Failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User logged in", "user_id", user.ID)

	resp := AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
