package handlers

import (
	"encoding/json"
	"godo/internal/auth"
	"godo/internal/domain"
	"godo/internal/service"
	"log/slog"
	"net/http"
	"time"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
	jwtSecret   string
}

func NewAuthHandler(authService *service.AuthService, logger *slog.Logger, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		jwtSecret:   jwtSecret,
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
	User  domain.User `json:"user"`
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
	}

	user, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to register user", "error", err)
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

	writeJsonResponse(w, http.StatusCreated, resp, h.logger)
}
