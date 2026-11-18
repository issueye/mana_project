package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mana/internal/model"
)

// ListComments returns one owner's comments, newest first.
func (s *Store) ListComments(ctx context.Context, ownerType string, ownerID int64) ([]model.Comment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, owner_type, owner_id, content, created_at, updated_at
		FROM comments WHERE owner_type = ? AND owner_id = ?
		ORDER BY created_at DESC, id DESC`, ownerType, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	return scanComments(rows)
}

// ListAllComments returns every comment, for the snapshot.
func (s *Store) ListAllComments(ctx context.Context) ([]model.Comment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, owner_type, owner_id, content, created_at, updated_at
		FROM comments
		ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	return scanComments(rows)
}

func scanComments(rows *sql.Rows) ([]model.Comment, error) {
	var out []model.Comment
	for rows.Next() {
		var c model.Comment
		var created, updated int64
		if err := rows.Scan(&c.ID, &c.OwnerType, &c.OwnerID, &c.Content, &created, &updated); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		c.CreatedAt = time.Unix(created, 0)
		c.UpdatedAt = time.Unix(updated, 0)
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateComment inserts one remark.
func (s *Store) CreateComment(ctx context.Context, ownerType string, ownerID int64, content string) (model.Comment, error) {
	now := s.now()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO comments(owner_type, owner_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		ownerType, ownerID, content, now.Unix(), now.Unix())
	if err != nil {
		return model.Comment{}, fmt.Errorf("create comment: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetComment(ctx, id)
}

// GetComment loads one remark.
func (s *Store) GetComment(ctx context.Context, id int64) (model.Comment, error) {
	var c model.Comment
	var created, updated int64
	err := s.db.QueryRowContext(ctx, `
		SELECT id, owner_type, owner_id, content, created_at, updated_at
		FROM comments WHERE id = ?`, id).
		Scan(&c.ID, &c.OwnerType, &c.OwnerID, &c.Content, &created, &updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Comment{}, ErrNotFound
		}
		return model.Comment{}, fmt.Errorf("get comment: %w", err)
	}
	c.CreatedAt = time.Unix(created, 0)
	c.UpdatedAt = time.Unix(updated, 0)
	return c, nil
}

// DeleteComment removes one remark.
func (s *Store) DeleteComment(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM comments WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// deleteCommentsFor removes every comment of one owner inside a host's
// delete transaction.
func deleteCommentsFor(tx *sql.Tx, ownerType string, ownerID int64) error {
	_, err := tx.ExecContext(context.Background(),
		`DELETE FROM comments WHERE owner_type = ? AND owner_id = ?`, ownerType, ownerID)
	return err
}
