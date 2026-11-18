-- F13 attachments: documents may carry an uploaded file. Notes keep the
-- empty defaults.

ALTER TABLE documents ADD COLUMN file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE documents ADD COLUMN file_size INTEGER NOT NULL DEFAULT 0;
