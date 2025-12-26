package handlers

import (
	"godo/gen/todo/v1/todov1connect"
	"godo/internal/service"
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
