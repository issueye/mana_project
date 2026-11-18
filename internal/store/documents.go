package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"mana/internal/model"
)

const documentSelect = `
	SELECT d.id, d.owner_type, d.owner_id,
	       CASE d.owner_type
	         WHEN 'project' THEN COALESCE((SELECT name FROM projects WHERE id = d.owner_id), '未命名项目')
	         ELSE COALESCE((SELECT name FROM project_groups WHERE id = d.owner_id), '未命名项目组')
	       END,
	       d.title, d.kind, d.link, length(d.content), d.created_at, d.updated_at,
	       d.file_name, d.file_size
	FROM documents d`

// ListDocuments lists metadata for one owner, or every document when
// ownerType is empty.
func (s *Store) ListDocuments(ctx context.Context, ownerType string, ownerID int64) ([]model.Document, error) {
	query := documentSelect
	var args []any
	if ownerType != "" {
		query += ` WHERE d.owner_type = ? AND d.owner_id = ?`
		args = append(args, ownerType, ownerID)
	}
	query += ` ORDER BY d.updated_at DESC, d.title COLLATE NOCASE`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()
	var out []model.Document
	for rows.Next() {
		d, err := scanDocumentMeta(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetDocument loads one document including its content.
func (s *Store) GetDocument(ctx context.Context, id int64) (model.Document, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT d.id, d.owner_type, d.owner_id,
		       CASE d.owner_type
		         WHEN 'project' THEN COALESCE((SELECT name FROM projects WHERE id = d.owner_id), '未命名项目')
		         ELSE COALESCE((SELECT name FROM project_groups WHERE id = d.owner_id), '未命名项目组')
		       END,
		       d.title, d.kind, d.content, d.link, length(d.content), d.created_at, d.updated_at,
		       d.file_name, d.file_size
		FROM documents d WHERE d.id = ?`, id)
	var d model.Document
	var created, updated int64
	if err := row.Scan(&d.ID, &d.OwnerType, &d.OwnerID, &d.OwnerName,
		&d.Title, &d.Kind, &d.Content, &d.Link, &d.Words, &created, &updated,
		&d.FileName, &d.FileSize); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Document{}, ErrNotFound
		}
		return model.Document{}, fmt.Errorf("get document: %w", err)
	}
	d.CreatedAt = time.Unix(created, 0)
	d.UpdatedAt = time.Unix(updated, 0)
	return d, nil
}

// CreateDocument inserts a document.
func (s *Store) CreateDocument(ctx context.Context, in model.DocumentInput) (model.Document, error) {
	now := s.now()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO documents(owner_type, owner_id, title, kind, content, link, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		in.OwnerType, in.OwnerID, in.Title, in.Kind, in.Content, in.Link,
		now.Unix(), now.Unix())
	if err != nil {
		return model.Document{}, mapError("create document", err)
	}
	id, _ := res.LastInsertId()
	return s.GetDocument(ctx, id)
}

// UpdateDocument updates a document's editable fields. The attachment
// columns stay as they are: a file's body does not change through the
// note editor.
func (s *Store) UpdateDocument(ctx context.Context, id int64, in model.DocumentInput) (model.Document, error) {
	now := s.now()
	res, err := s.db.ExecContext(ctx, `
		UPDATE documents
		SET title = ?, kind = ?, content = ?, link = ?, updated_at = ?
		WHERE id = ?`,
		in.Title, in.Kind, in.Content, in.Link, now.Unix(), id)
	if err != nil {
		return model.Document{}, mapError("update document", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.Document{}, ErrNotFound
	}
	return s.GetDocument(ctx, id)
}

// DeleteDocument removes one document.
func (s *Store) DeleteDocument(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// CountDocuments counts a given owner's documents.
func (s *Store) CountDocuments(ctx context.Context, ownerType string, ownerID int64) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents WHERE owner_type = ? AND owner_id = ?`,
		ownerType, ownerID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return n, nil
}

// SetFilesDir points the store at the directory that holds uploaded
// attachments; empty disables file uploads.
func (s *Store) SetFilesDir(dir string) { s.filesDir = dir }

// FilesDir reports the attachments root, empty when unset.
func (s *Store) FilesDir() string { return s.filesDir }

// AttachmentPath is where a document's file lives, relative to the files
// dir.
func AttachmentPath(docID int64, fileName string) string {
	return filepath.Join("attachments", strconv.FormatInt(docID, 10), fileName)
}

// SaveFile inserts a file document and copies the uploaded file into the
// attachments tree. The extracted preview text travels in in.Content.
// A failed copy removes the new row again.
func (s *Store) SaveFile(ctx context.Context, in model.DocumentInput, srcPath, fileName string) (model.Document, error) {
	if s.filesDir == "" {
		return model.Document{}, fmt.Errorf("save file: attachments directory is not configured")
	}
	if fileName == "" {
		return model.Document{}, fmt.Errorf("save file: file name is empty")
	}
	now := s.now()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO documents(owner_type, owner_id, title, kind, content, link, file_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.OwnerType, in.OwnerID, in.Title, in.Kind, in.Content, in.Link, fileName,
		now.Unix(), now.Unix())
	if err != nil {
		return model.Document{}, mapError("create document", err)
	}
	id, _ := res.LastInsertId()

	rel := AttachmentPath(id, fileName)
	dest := filepath.Join(s.filesDir, rel)
	if err := copyFile(srcPath, dest); err != nil {
		s.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
		os.RemoveAll(filepath.Dir(dest))
		return model.Document{}, fmt.Errorf("save file: %w", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		return model.Document{}, fmt.Errorf("save file: %w", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE documents SET file_size = ? WHERE id = ?`, info.Size(), id); err != nil {
		return model.Document{}, fmt.Errorf("save file: %w", err)
	}
	return s.GetDocument(ctx, id)
}

// RemoveAttachment deletes a document's copied file tree, best effort.
func (s *Store) RemoveAttachment(docID int64) {
	if s.filesDir == "" {
		return
	}
	os.RemoveAll(filepath.Join(s.filesDir, "attachments", strconv.FormatInt(docID, 10)))
}

// AttachmentFile returns the absolute path of a stored attachment.
func (s *Store) AttachmentFile(docID int64, fileName string) (string, error) {
	if s.filesDir == "" || fileName == "" {
		return "", fmt.Errorf("document has no attached file")
	}
	path := filepath.Join(s.filesDir, AttachmentPath(docID, fileName))
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("attachment is missing: %w", err)
	}
	return path, nil
}

// copyFile copies src to dest, creating the parent directories.
func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
