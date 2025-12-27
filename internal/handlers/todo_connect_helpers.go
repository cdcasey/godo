package handlers

import (
	todov1 "godo/gen/todo/v1"
	"godo/internal/domain"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// This may seem weird, but here's the deal:
// We can't use domain.Todo in the protobuf code, because protobuf
// files aren't Go code. If we used the generataed todov1.Todo in the domain,
// then the domain becomes dependent on the API contract without the ability
// to be it's own thing. With this function, We can create a new API schema (v2)
// without necessarily having to update the domain, and we don't have to worry
// about surfacing fields we don't necessarily want to in the API.
// domain.Todo == business view of a todo
// todov1.Todo == API view of a todo
func todoToProto(t *domain.Todo) *todov1.Todo {
	return &todov1.Todo{
		Id:          t.ID,
		UserId:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}
