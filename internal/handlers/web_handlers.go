package handlers

import (
	"fmt"
	"net/http"
	"time"

	"godo/internal/auth"
	"godo/internal/service"
	"godo/web/templates/pages"
)

type WebHandler struct {
	authService *service.AuthService
	jwtSecret   string
}

func NewWebHandler(authService *service.AuthService, jwtSecret string) *WebHandler {
	return &WebHandler{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

func (h *WebHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.Login().Render(r.Context(), w)
}

func (h *WebHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.authService.Authenticate(email, password)
	if err != nil {
		fmt.Println(fmt.Errorf("the error: %w", err))
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Invalid email or password"))
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, h.jwtSecret, 24*time.Hour)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Something went wrong"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	w.Header().Set("HX-Redirect", "/todos")
	w.WriteHeader(http.StatusOK)
}
