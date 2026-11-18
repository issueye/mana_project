package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"mana/internal/model"
)

const projectSelect = `
	SELECT p.id,
	       p.group_id,
	       COALESCE(g.name, '未分组'),
	       p.name,
	       p.tags,
	       p.description,
	       p.path,
	       p.version,
	       p.created_at,
	       p.updated_at,
	       (SELECT COUNT(*) FROM documents d
	          WHERE d.owner_type = 'project' AND d.owner_id = p.id)
	FROM projects p
	LEFT JOIN project_groups g ON g.id = p.group_id`

// ListProjects returns all projects ordered by update time.
func (s *Store) ListProjects(ctx context.Context) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, projectSelect+`
		ORDER BY p.updated_at DESC, p.name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var out []model.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetProject returns one project by id.
func (s *Store) GetProject(ctx context.Context, id int64) (model.Project, error) {
	row := s.db.QueryRowContext(ctx, projectSelect+` WHERE p.id = ?`, id)
	p, err := scanProjectRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Project{}, ErrNotFound
		}
		return model.Project{}, err
	}
	return p, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProject(rows *sql.Rows) (model.Project, error) {
	return scanProjectRow(rows)
}

func scanProjectRow(row rowScanner) (model.Project, error) {
	var p model.Project
	var groupID sql.NullInt64
	var tags string
	var created, updated int64
	if err := row.Scan(&p.ID, &groupID, &p.GroupName, &p.Name, &tags, &p.Description,
		&p.Path, &p.Version, &created, &updated, &p.DocumentCount); err != nil {
		return model.Project{}, fmt.Errorf("scan project: %w", err)
	}
	if groupID.Valid {
		v := groupID.Int64
		p.GroupID = &v
	}
	p.Tags = decodeTags(tags)
	p.CreatedAt = time.Unix(created, 0)
	p.UpdatedAt = time.Unix(updated, 0)
	return p, nil
}

func encodeTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func decodeTags(raw string) []string {
	if raw == "" {
		return nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil
	}
	return tags
}

// CreateProject inserts a project.
func (s *Store) CreateProject(ctx context.Context, in model.ProjectInput) (model.Project, error) {
	now := s.now()
	var groupID any
	if in.GroupID != nil {
		groupID = *in.GroupID
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO projects(group_id, name, tags, description, path, version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		groupID, in.Name, encodeTags(in.Tags), in.Description, in.Path, in.Version,
		now.Unix(), now.Unix())
	if err != nil {
		return model.Project{}, mapError("create project", err)
	}
	id, _ := res.LastInsertId()
	return s.GetProject(ctx, id)
}

// UpdateProject updates a project and optionally moves it to another group.
func (s *Store) UpdateProject(ctx context.Context, id int64, in model.ProjectInput) (model.Project, error) {
	now := s.now()
	var groupID any
	if in.GroupID != nil {
		groupID = *in.GroupID
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE projects
		SET group_id = ?, name = ?, tags = ?, description = ?, path = ?, version = ?, updated_at = ?
		WHERE id = ?`,
		groupID, in.Name, encodeTags(in.Tags), in.Description, in.Path, in.Version,
		now.Unix(), id)
	if err != nil {
		return model.Project{}, mapError("update project", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.Project{}, ErrNotFound
	}
	return s.GetProject(ctx, id)
}

// DeleteProject deletes the project and its documents in one transaction.
func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete project: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM documents WHERE owner_type = 'project' AND owner_id = ?`, id); err != nil {
		return fmt.Errorf("delete project docs: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}

// MoveProject changes a project's group; nil moves it to ungrouped.
func (s *Store) MoveProject(ctx context.Context, id int64, groupID *int64) error {
	var gid any
	if groupID != nil {
		gid = *groupID
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE projects SET group_id = ?, updated_at = ? WHERE id = ?`,
		gid, s.now().Unix(), id)
	if err != nil {
		return mapError("move project", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ProjectDocuments returns the metadata of a project's documents.
func (s *Store) ProjectDocuments(ctx context.Context, projectID int64) ([]model.Document, error) {
	return s.ListDocuments(ctx, model.OwnerProject, projectID)
}
