package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func pointToString(s string) *string { return &s }
func pointToBool(b bool) *bool       { return &b }

func TestCreateTodo(t *testing.T) {
	model := NewMemoryModel()
	model.CreateTodo(ActionToSync{
		Title:     pointToString("t"),
		Completed: pointToBool(true),
	})
	assert.Equal(t, []Todo{{
		Id:        1,
		Title:     "t",
		Completed: true,
	}}, model.Todos)
	assert.Equal(t, 2, model.NextTodoId)
}
