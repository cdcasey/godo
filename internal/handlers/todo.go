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
	Completed   *bool   `json:"completed,omitempty"`
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
		return
	}

	h.logger.Info("Todo created", "todo_id", todo.ID, "user_id", claims.UserID)

	h.respondJson(w, http.StatusCreated, TodoResponse{Todo: *todo})
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

	if claims.Role == models.RoleAdmin {
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

	h.respondJson(w, http.StatusOK, TodosResponse{Todos: todos})
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
		return
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

	h.respondJson(w, http.StatusOK, TodoResponse{Todo: *todo})
}

// Handler to update a todo
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	// Claude's feedback on abstracting this logic out. Interesting:
	// My honest take: This is borderline over-engineering for two call sites. But if you expect more handlers to need
	// this pattern, go for it.
	todo, err := h.store.GetTodoByID(todoID)
	if err != nil {
		if errors.Is(err, store.ErrTodoNotFound) {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get todo", "error", err, "todo_id", todoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Users can only update their own todos
	if claims.Role != models.RoleAdmin && todo.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	todo.UpdatedAt = time.Now()

	if err := h.store.UpdateTodo(todo); err != nil {
		h.logger.Error("Failed to update todo", "error", err, "todo_id", todoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Todo updated", "todo_id", todoID, "user_id", claims.UserID)

	h.respondJson(w, http.StatusOK, TodoResponse{Todo: *todo})
}

// Handler to delete a todo
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Admin only
	if claims.Role != models.RoleAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	todoID := chi.URLParam(r, "id")
	if todoID == "" {
		http.Error(w, "Todo ID required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteTodo(todoID); err != nil {
		if errors.Is(err, store.ErrTodoNotFound) {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to delete todo", "error", err, "todo_id", todoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Todo deleted", "todo_id", todoID, "admin_id", claims.UserID)

	w.WriteHeader(http.StatusNoContent)
}
