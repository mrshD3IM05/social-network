-- Re-add post type column (best-effort rollback)
ALTER TABLE posts ADD COLUMN type TEXT NOT NULL DEFAULT 'post';