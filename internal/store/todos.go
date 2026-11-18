package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mana/internal/model"
)

// ListTodos returns a project's todos, open ones first, both groups
// newest-first.
func (s *Store) ListTodos(ctx context.Context, projectID int64) ([]model.Todo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, content, done, created_at, updated_at
		FROM todos WHERE project_id = ?
		ORDER BY done ASC, updated_at DESC, id DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()
	return scanTodos(rows)
}

// ListAllTodos returns every todo, for the snapshot.
func (s *Store) ListAllTodos(ctx context.Context) ([]model.Todo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, content, done, created_at, updated_at
		FROM todos
		ORDER BY done ASC, updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()
	return scanTodos(rows)
}

func scanTodos(rows *sql.Rows) ([]model.Todo, error) {
	var out []model.Todo
	for rows.Next() {
		var t model.Todo
		var done, created, updated int64
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Content, &done, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		t.Done = done != 0
		t.CreatedAt = time.Unix(created, 0)
		t.UpdatedAt = time.Unix(updated, 0)
		out = append(out, t)
	}
	return out, rows.Err()
}

// CreateTodo inserts one item.
func (s *Store) CreateTodo(ctx context.Context, projectID int64, content string) (model.Todo, error) {
	now := s.now()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO todos(project_id, content, done, created_at, updated_at)
		VALUES (?, ?, 0, ?, ?)`,
		projectID, content, now.Unix(), now.Unix())
	if err != nil {
		return model.Todo{}, fmt.Errorf("create todo: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetTodo(ctx, id)
}

// GetTodo loads one item.
func (s *Store) GetTodo(ctx context.Context, id int64) (model.Todo, error) {
	var t model.Todo
	var done, created, updated int64
	err := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, content, done, created_at, updated_at
		FROM todos WHERE id = ?`, id).
		Scan(&t.ID, &t.ProjectID, &t.Content, &done, &created, &updated)
	if err != nil {
		return model.Todo{}, mapError("get todo", err)
	}
	t.Done = done != 0
	t.CreatedAt = time.Unix(created, 0)
	t.UpdatedAt = time.Unix(updated, 0)
	return t, nil
}

// SetTodoDone toggles an item's done flag.
func (s *Store) SetTodoDone(ctx context.Context, id int64, done bool) (model.Todo, error) {
	now := s.now()
	doneInt := int64(0)
	if done {
		doneInt = 1
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE todos SET done = ?, updated_at = ? WHERE id = ?`, doneInt, now.Unix(), id)
	if err != nil {
		return model.Todo{}, fmt.Errorf("update todo: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.Todo{}, ErrNotFound
	}
	return s.GetTodo(ctx, id)
}

// DeleteTodo removes one item.
func (s *Store) DeleteTodo(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ClearDoneTodos removes every finished item of a project and reports
// how many went away.
func (s *Store) ClearDoneTodos(ctx context.Context, projectID int64) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM todos WHERE project_id = ? AND done = 1`, projectID)
	if err != nil {
		return 0, fmt.Errorf("clear todos: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
