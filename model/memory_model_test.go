package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func pointToString(s string) *string { return &s }
func pointToBool(b bool) *bool       { return &b }

func TestCreateTodo(t *testing.T) {
	model := NewMemoryModel()
	spec := struct {
		Title     string
		Completed bool
	}{"t", true}
	newTodo := model.CreateTodo(ActionToSync{
		Title:     &spec.Title,
		Completed: &spec.Completed,
	})
	assert.Equal(t, spec.Title, newTodo.Title)
	assert.Equal(t, spec.Completed, newTodo.Completed)
	assert.Equal(t, []Todo{{
		Id:        1,
		Title:     spec.Title,
		Completed: spec.Completed,
	}}, model.Todos)
	assert.Equal(t, 2, model.NextTodoId)
}
