-- Mana v1.0 schema.
-- PRAGMA statements run outside this migration, before it is applied.

CREATE TABLE project_groups (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL COLLATE NOCASE,
    description TEXT    NOT NULL DEFAULT '',
    color       TEXT    NOT NULL DEFAULT '#7C5CFF',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE UNIQUE INDEX ux_project_groups_name ON project_groups(name);

CREATE TABLE projects (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id    INTEGER REFERENCES project_groups(id) ON DELETE SET NULL,
    name        TEXT    NOT NULL,
    tags        TEXT    NOT NULL DEFAULT '[]',
    description TEXT    NOT NULL DEFAULT '',
    path        TEXT    NOT NULL DEFAULT '',
    version     TEXT    NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE UNIQUE INDEX ux_projects_group_name
    ON projects(IFNULL(group_id, 0), name COLLATE NOCASE);
CREATE INDEX ix_projects_group ON projects(group_id);
CREATE INDEX ix_projects_updated ON projects(updated_at DESC);

CREATE TABLE documents (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_type TEXT    NOT NULL CHECK (owner_type IN ('project','group')),
    owner_id   INTEGER NOT NULL,
    title      TEXT    NOT NULL,
    kind       TEXT    NOT NULL DEFAULT 'other',
    content    TEXT    NOT NULL DEFAULT '',
    link       TEXT    NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE UNIQUE INDEX ux_documents_owner_title
    ON documents(owner_type, owner_id, title COLLATE NOCASE);
CREATE INDEX ix_documents_owner ON documents(owner_type, owner_id, updated_at DESC);
