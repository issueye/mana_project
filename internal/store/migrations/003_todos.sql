-- F17 todos: per-project action items. Deleting a project takes its
-- todos along (foreign_keys pragma is on).

CREATE TABLE todos (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    content    TEXT    NOT NULL,
    done       INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX ix_todos_project ON todos(project_id, done, updated_at DESC);
