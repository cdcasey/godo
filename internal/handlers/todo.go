package handlers

import (
	"encoding/json"
	"errors"
	"godo/internal/auth"
	"godo/internal/models"
	"godo/internal/store"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TodoHandler struct {
	store  store.TodoStore
	logger *slog.Logger
}

func NewTodoHandler(store store.TodoStore, logger *slog.Logger) *TodoHandler {
	return &TodoHandler{
		store:  store,
		logger: logger,
	}
}

type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Note: UpdateTodoRequest uses pointers to distinguish between "not provided" and "set to empty/false".
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *string `json:"completed,omitempty"`
}

type TodoResponse struct {
	Todo models.Todo `json:"todo"`
}

type TodosResponse struct {
	Todos []*models.Todo `json:"todos"`
}

// Handler for creating a new todo
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo := models.NewTodo(claims.UserID, req.Title, req.Description)

	if err := h.store.CreateTodo(todo); err != nil {
		h.logger.Error("Failed to create todo", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	h.logger.Info("Todo created", "todo_id", todo.ID, "user_id", claims.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TodoResponse{Todo: *todo})
}

// Handler for listing all todos
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var todos []*models.Todo
	var err error

	if claims.Role == "admin" {
		todos, err = h.store.GetAllTodos()
	} else {
		todos, err = h.store.GetTodosByUserID(claims.UserID)
	}

	if err != nil {
		h.logger.Error("Failed to get todos", "error", err, "user_id", claims.UserID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Todos listed", "user_id", claims.UserID, "count", len(todos))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TodosResponse{Todos: todos})
}

// Handler for getting a single todo
func (h *TodoHandler) GetById(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	todoID := chi.URLParam(r, "id")
	if todoID == "" {
		http.Error(w, "Todo ID required", http.StatusBadRequest)
	}

	todo, err := h.store.GetTodoByID(todoID)
	if err != nil {
		// if err == store.ErrTodoNotFound {
		if errors.Is(err, store.ErrTodoNotFound) {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get todo", "error", err, "todo_id", todoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Users can only view their own todos, admins can view any
	if claims.Role != models.RoleAdmin && todo.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "Application/json")
	json.NewEncoder(w).Encode(TodoResponse{Todo: *todo})
}
