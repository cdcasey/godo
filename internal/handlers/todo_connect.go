package handlers

import (
	"context"
	"errors"
	todov1 "godo/gen/todo/v1"
	"godo/gen/todo/v1/todov1connect"
	"godo/internal/auth"
	"godo/internal/domain"
	"godo/internal/service"

	"connectrpc.com/connect"
)

var _ todov1connect.TodoServiceHandler = (*TodoConnectHandler)(nil)

type TodoConnectHandler struct {
	todoService *service.TodoService
}

func NewTodoConnectHandler(todoService *service.TodoService) *TodoConnectHandler {
	return &TodoConnectHandler{
		todoService: todoService,
	}
}

func (h *TodoConnectHandler) CreateTodo(
	ctx context.Context,
	req *connect.Request[todov1.CreateTodoRequest],
) (*connect.Response[todov1.CreateTodoResponse], error) {
	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	todo, err := h.todoService.Create(
		claims.UserID,
		req.Msg.Title,
		req.Msg.Description,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&todov1.CreateTodoResponse{
		Todo: todoToProto(todo),
	}), nil
}

func (h *TodoConnectHandler) ListTodos(
	ctx context.Context,
	req *connect.Request[todov1.ListTodosRequest],
) (*connect.Response[todov1.ListTodosResponse], error) {
	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	todos, err := h.todoService.List(claims.UserID, claims.Role)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protos := make([]*todov1.Todo, len(todos))
	for i, todo := range todos {
		protos[i] = todoToProto(todo)
	}

	return connect.NewResponse(&todov1.ListTodosResponse{
		Todos: protos,
	}), nil
}

func (h *TodoConnectHandler) GetTodo(
	ctx context.Context, req *connect.Request[todov1.GetTodoRequest],
) (*connect.Response[todov1.GetTodoResponse], error) {
	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	// Thanks to protobuf typing this probably isn't needed, but I'm keeping it as an example
	todoID := req.Msg.Id
	if todoID == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("Todo ID required"))
	}

	todo, err := h.todoService.GetByID(todoID, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, nil)
		}
		if errors.Is(err, service.ErrForbidden) {
			return nil, connect.NewError(connect.CodePermissionDenied, nil)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&todov1.GetTodoResponse{Todo: todoToProto(todo)}), nil
}

func (h *TodoConnectHandler) UpdateTodo(ctx context.Context, req *connect.Request[todov1.UpdateTodoRequest]) (*connect.Response[todov1.UpdateTodoResponse], error) {
	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	// This time let's assume that protobuf typing worked
	todo, err := h.todoService.Update(req.Msg.Id, claims.UserID, claims.Role, req.Msg.Title, req.Msg.Description, req.Msg.Completed)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, nil)
		}
		if errors.Is(err, service.ErrForbidden) {
			return nil, connect.NewError(connect.CodePermissionDenied, nil)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&todov1.UpdateTodoResponse{Todo: todoToProto(todo)}), nil
}

func (h *TodoConnectHandler) DeleteTodo(ctx context.Context, req *connect.Request[todov1.DeleteTodoRequest]) (*connect.Response[todov1.DeleteTodoResponse], error) {
	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	err := h.todoService.Delete(req.Msg.Id, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, domain.ErrTodoNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, nil)
		}
		if errors.Is(err, service.ErrForbidden) {
			return nil, connect.NewError(connect.CodePermissionDenied, nil)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&todov1.DeleteTodoResponse{}), nil
}
