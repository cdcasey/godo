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

type TodoHandler struct {
	todoService *service.TodoService
	logger      *slog.Logger
}

func NewTodoHandler(todoService *service.TodoService, logger *slog.Logger) *TodoHandler {
	return &TodoHandler{
		todoService: todoService,
		logger:      logger,
	}
}

type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

type TodoResponse struct {
	Todo domain.Todo `json:"todo"`
}

type TodosResponse struct {
	Todos []*domain.Todo `json:"todos"`
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

	todo, err := h.todoService.Create(claims.UserID, req.Title, req.Description)
	if err != nil {
		h.logger.Error("Failed to create todo", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Todo created", "todo_id", todo.ID, "user_id", claims.UserID)

	writeJsonResponse(w, http.StatusCreated, TodoResponse{Todo: *todo}, h.logger)
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	todos, err := h.todoService.List(claims.UserID, claims.Role)
	if err != nil {
		h.logger.Error("Failed to get todos", "error", err, "user_id", claims.UserID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Todos listed", "user_id", claims.UserID, "count", len(todos))

	writeJsonResponse(w, http.StatusOK, TodosResponse{Todos: todos}, h.logger)
}

func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	todo, err := h.todoService.GetByID(todoID, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to get todo", "error", err, "todo_id", todoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, http.StatusOK, TodoResponse{Todo: *todo}, h.logger)
}

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

	var req UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	todo, err := h.todoService.Update(todoID, claims.UserID, claims.Role, req.Title, req.Description, req.Completed)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to update todo", "error", err, "todo_id", todoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Todo updated", "todo_id", todoID, "user_id", claims.UserID)

	writeJsonResponse(w, http.StatusOK, TodoResponse{Todo: *todo}, h.logger)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	err := h.todoService.Delete(todoID, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("Failed to delete todo", "error", err, "todo_id", todoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Todo deleted", "todo_id", todoID, "user_id", claims.UserID)

	w.WriteHeader(http.StatusNoContent)
}
