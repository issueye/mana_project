package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mana/internal/model"
)

// Search matches a query against project names, paths and tags, document
// titles and group names. Every section is capped so the palette fills
// fast; the empty query returns an empty result.
func (s *Store) Search(ctx context.Context, q string) (model.SearchResult, error) {
	q = strings.TrimSpace(q)
	out := model.SearchResult{Query: q}
	if q == "" {
		return out, nil
	}
	like := "%" + q + "%"
	tagLike := `%"` + q + `"%`

	projectSQL := `
		SELECT p.id, p.group_id, IFNULL(g.name, ''), p.name, p.tags, p.description,
		       p.path, p.version, p.created_at, p.updated_at,
		       (SELECT COUNT(*) FROM documents d WHERE d.owner_type='project' AND d.owner_id = p.id)
		FROM projects p LEFT JOIN project_groups g ON g.id = p.group_id
		WHERE p.name LIKE ? COLLATE NOCASE
		   OR p.path LIKE ? COLLATE NOCASE
		   OR p.tags LIKE ? COLLATE NOCASE
		ORDER BY p.updated_at DESC
		LIMIT 5`
	rows, err := s.db.QueryContext(ctx, projectSQL, like, like, tagLike)
	if err != nil {
		return out, fmt.Errorf("search projects: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanProjectRow(rows)
		if err != nil {
			return out, err
		}
		out.Projects = append(out.Projects, p)
	}
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("search projects: %w", err)
	}

	docSQL := documentSelect + `
		WHERE d.title LIKE ? COLLATE NOCASE
		ORDER BY d.updated_at DESC
		LIMIT 5`
	rows, err = s.db.QueryContext(ctx, docSQL, like)
	if err != nil {
		return out, fmt.Errorf("search documents: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		d, err := scanDocumentMeta(rows)
		if err != nil {
			return out, err
		}
		out.Docs = append(out.Docs, d)
	}
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("search documents: %w", err)
	}

	groupSQL := `
		SELECT g.id, g.name, g.description, g.color, g.created_at, g.updated_at,
		       (SELECT COUNT(*) FROM projects p WHERE p.group_id = g.id),
		       (SELECT COUNT(*) FROM documents d WHERE d.owner_type='group' AND d.owner_id = g.id)
		FROM project_groups g
		WHERE g.name LIKE ? COLLATE NOCASE
		ORDER BY g.name COLLATE NOCASE
		LIMIT 5`
	rows, err = s.db.QueryContext(ctx, groupSQL, like)
	if err != nil {
		return out, fmt.Errorf("search groups: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		g, err := scanGroupRow(rows)
		if err != nil {
			return out, err
		}
		out.Groups = append(out.Groups, g)
	}
	err = rows.Err()
	if err != nil {
		err = fmt.Errorf("search groups: %w", err)
	}
	return out, err
}

// scanDocumentMeta scans the metadata columns ListDocuments uses, for the
// search path.
func scanDocumentMeta(row rowScanner) (model.Document, error) {
	var d model.Document
	var created, updated int64
	if err := row.Scan(&d.ID, &d.OwnerType, &d.OwnerID, &d.OwnerName,
		&d.Title, &d.Kind, &d.Link, &d.Words, &created, &updated,
		&d.FileName, &d.FileSize); err != nil {
		return d, fmt.Errorf("scan document: %w", err)
	}
	d.CreatedAt = time.Unix(created, 0)
	d.UpdatedAt = time.Unix(updated, 0)
	return d, nil
}

// scanGroupRow scans the group columns ListGroups uses.
func scanGroupRow(row rowScanner) (model.Group, error) {
	var g model.Group
	var created, updated int64
	if err := row.Scan(&g.ID, &g.Name, &g.Description, &g.Color, &created, &updated,
		&g.ProjectCount, &g.DocumentCount); err != nil {
		return g, fmt.Errorf("scan group: %w", err)
	}
	g.CreatedAt = time.Unix(created, 0)
	g.UpdatedAt = time.Unix(updated, 0)
	return g, nil
}
