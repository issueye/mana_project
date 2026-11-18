package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"mana/internal/model"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Store owns the SQLite connection. It is safe for concurrent use; the
// database/sql pool is pinned to one connection because the app is local
// and small.
type Store struct {
	db       *sql.DB
	now      func() time.Time
	filesDir string
}

// Open opens (and creates) the database at path and applies migrations.
// Use ":memory:" for tests.
func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("foreign keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, `PRAGMA busy_timeout = 5000`); err != nil {
		db.Close()
		return nil, fmt.Errorf("busy timeout: %w", err)
	}
	if path != ":memory:" {
		// WAL keeps the local database readable after an abnormal exit.
		if _, err := db.ExecContext(ctx, `PRAGMA journal_mode = WAL`); err != nil {
			db.Close()
			return nil, fmt.Errorf("journal mode: %w", err)
		}
	}
	s := &Store{db: db, now: func() time.Time { return time.Now().UTC() }}
	if err := s.Migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the connection for the migration tests and diagnostics.
func (s *Store) DB() *sql.DB { return s.db }

// Migrate applies embedded migrations once each.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`); err != nil {
		return fmt.Errorf("migrations table: %w", err)
	}
	entries, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)
	for _, name := range entries {
		version, err := migrationVersion(name)
		if err != nil {
			return err
		}
		var count int
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", version, err)
		}
		if count > 0 {
			continue
		}
		body, err := migrationFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		statements := splitSQL(string(body))
		if len(statements) == 0 {
			continue
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", version, err)
		}
		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d: %w", version, err)
			}
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`,
			version, s.now().Unix()); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}
	}
	return nil
}

func migrationVersion(name string) (int, error) {
	base := strings.TrimPrefix(name, "migrations/")
	base = strings.TrimSuffix(base, ".sql")
	part := strings.SplitN(base, "_", 2)[0]
	v, err := strconv.Atoi(part)
	if err != nil {
		return 0, fmt.Errorf("migration name %q: %w", name, err)
	}
	return v, nil
}

// splitSQL splits a migration file into single statements. The migrations
// contain no string literals with semicolons; line comments are removed.
func splitSQL(src string) []string {
	var clean strings.Builder
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		clean.WriteString(line)
		clean.WriteByte('\n')
	}
	var out []string
	for _, part := range strings.Split(clean.String(), ";") {
		if statement := strings.TrimSpace(part); statement != "" {
			out = append(out, statement)
		}
	}
	return out
}

// Snapshot loads the complete metadata view used by the UI.
func (s *Store) Snapshot(ctx context.Context) (model.Snapshot, error) {
	groups, err := s.ListGroups(ctx)
	if err != nil {
		return model.Snapshot{}, err
	}
	projects, err := s.ListProjects(ctx)
	if err != nil {
		return model.Snapshot{}, err
	}
	docs, err := s.ListDocuments(ctx, "", 0)
	if err != nil {
		return model.Snapshot{}, err
	}
	return model.Snapshot{Groups: groups, Projects: projects, Docs: docs}, nil
}

// ListGroups returns all groups with their project and document counts.
func (s *Store) ListGroups(ctx context.Context) ([]model.Group, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.description, g.color, g.created_at, g.updated_at,
		       (SELECT COUNT(*) FROM projects p WHERE p.group_id = g.id),
		       (SELECT COUNT(*) FROM documents d
		          WHERE d.owner_type = 'group' AND d.owner_id = g.id)
		FROM project_groups g
		ORDER BY g.created_at, g.id`)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()
	var groups []model.Group
	for rows.Next() {
		var g model.Group
		var created, updated int64
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Color, &created, &updated,
			&g.ProjectCount, &g.DocumentCount); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		g.CreatedAt = time.Unix(created, 0)
		g.UpdatedAt = time.Unix(updated, 0)
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// GetGroup loads one group by id.
func (s *Store) GetGroup(ctx context.Context, id int64) (model.Group, error) {
	groups, err := s.ListGroups(ctx)
	if err != nil {
		return model.Group{}, err
	}
	for _, g := range groups {
		if g.ID == id {
			return g, nil
		}
	}
	return model.Group{}, ErrNotFound
}

// CreateGroup inserts a group.
func (s *Store) CreateGroup(ctx context.Context, in model.GroupInput) (model.Group, error) {
	now := s.now()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO project_groups(name, description, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		in.Name, in.Description, in.Color, now.Unix(), now.Unix())
	if err != nil {
		return model.Group{}, mapError("create group", err)
	}
	id, _ := res.LastInsertId()
	return s.GetGroup(ctx, id)
}

// UpdateGroup updates a group.
func (s *Store) UpdateGroup(ctx context.Context, id int64, in model.GroupInput) (model.Group, error) {
	now := s.now()
	res, err := s.db.ExecContext(ctx, `
		UPDATE project_groups SET name = ?, description = ?, color = ?, updated_at = ?
		WHERE id = ?`,
		in.Name, in.Description, in.Color, now.Unix(), id)
	if err != nil {
		return model.Group{}, mapError("update group", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.Group{}, ErrNotFound
	}
	return s.GetGroup(ctx, id)
}

// DeleteGroup removes a group, moves its projects to ungrouped, and deletes
// its group documents in one transaction.
func (s *Store) DeleteGroup(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete group: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`UPDATE projects SET group_id = NULL WHERE group_id = ?`, id); err != nil {
		return fmt.Errorf("move projects: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM documents WHERE owner_type = 'group' AND owner_id = ?`, id); err != nil {
		return fmt.Errorf("delete group docs: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM project_groups WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}
