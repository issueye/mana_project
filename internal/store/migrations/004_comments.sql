-- F18 comments: a generic comment stream for projects, groups and
-- documents. The owner is cleaned up in the host's delete transaction,
-- since one table cannot carry three foreign keys.

CREATE TABLE comments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_type TEXT    NOT NULL CHECK (owner_type IN ('project','group','document')),
    owner_id   INTEGER NOT NULL,
    content    TEXT    NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX ix_comments_owner ON comments(owner_type, owner_id, created_at DESC);
