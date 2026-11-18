package store

import (
	"context"
	"testing"

	"mana/internal/model"
)

func newTodoProject(t *testing.T, s *Store, ctx context.Context) int64 {
	t.Helper()
	p, err := s.CreateProject(ctx, model.ProjectInput{Name: "待办项目"})
	if err != nil {
		t.Fatal(err)
	}
	return p.ID
}

func TestTodosLifecycle(t *testing.T) {
	s, ctx := testStore(t)
	projectID := newTodoProject(t, s, ctx)

	first, err := s.CreateTodo(ctx, projectID, "写发布说明")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateTodo(ctx, projectID, "联调上传")
	if err != nil {
		t.Fatal(err)
	}

	todos, err := s.ListTodos(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 2 || todos[0].Content != "联调上传" {
		t.Fatalf("order = %#v", todos)
	}

	if _, err := s.SetTodoDone(ctx, second.ID, true); err != nil {
		t.Fatal(err)
	}
	todos, _ = s.ListTodos(ctx, projectID)
	if !todos[1].Done || todos[0].Done {
		t.Fatalf("done ordering = %#v", todos)
	}

	if err := s.DeleteTodo(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	todos, _ = s.ListTodos(ctx, projectID)
	if len(todos) != 1 {
		t.Fatalf("after delete = %#v", todos)
	}

	n, err := s.ClearDoneTodos(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("cleared = %d", n)
	}
	if todos, _ = s.ListTodos(ctx, projectID); len(todos) != 0 {
		t.Fatalf("clear left %#v", todos)
	}
}

// Deleting a project cascades to its todos.
func TestTodosCascadeWithProject(t *testing.T) {
	s, ctx := testStore(t)
	projectID := newTodoProject(t, s, ctx)
	if _, err := s.CreateTodo(ctx, projectID, "会级联的待办"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteProject(ctx, projectID); err != nil {
		t.Fatal(err)
	}
	todos, err := s.ListTodos(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 0 {
		t.Fatalf("todos survived project delete: %#v", todos)
	}
	all, err := s.ListAllTodos(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Fatalf("snapshot todos survived: %#v", all)
	}
}
