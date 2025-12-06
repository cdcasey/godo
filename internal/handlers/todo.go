package handlers

import (
	"encoding/json"
	"godo/internal/auth"
	"godo/internal/models"
	"godo/internal/store"
	"log/slog"
	"net/http"
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
type UpdateTodosRequest struct {
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
