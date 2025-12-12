package handlers

import (
	"net/http"
	"time"

	"godo/internal/auth"
	"godo/internal/service"
	"godo/web/templates/components"
	"godo/web/templates/pages"

	"github.com/go-chi/chi/v5"
)

type WebHandler struct {
	authService *service.AuthService
	todoService *service.TodoService
	jwtSecret   string
}

func NewWebHandler(authService *service.AuthService, todoService *service.TodoService, jwtSecret string) *WebHandler {
	return &WebHandler{
		authService: authService,
		todoService: todoService,
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

func (h *WebHandler) TodosPage(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	todos, err := h.todoService.List(claims.UserID, claims.Role)
	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	pages.Todos(todos).Render(r.Context(), w)
}

func (h *WebHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo, err := h.todoService.Create(claims.UserID, title, "")
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	components.TodoItem(todo).Render(r.Context(), w)
}

func (h *WebHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	todoID := chi.URLParam(r, "id")
	if todoID == "" {
		http.Error(w, "Todo ID required", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	completedStr := r.FormValue("completed")
	completed := completedStr == "true"

	todo, err := h.todoService.Update(todoID, claims.UserID, claims.Role, nil, nil, &completed)
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	components.TodoItem(todo).Render(r.Context(), w)
}
