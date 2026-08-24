DROP INDEX IF EXISTS idx_files_message;

-- SQLite does not support DROP COLUMN before 3.35; recreate table without the column.
CREATE TABLE IF NOT EXISTS files_backup (
    id           TEXT PRIMARY KEY,
    storage_path TEXT NOT NULL UNIQUE,
    original_name TEXT NOT NULL,
    mime_type    TEXT NOT NULL,
    size         INTEGER NOT NULL,
    owner_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id      INTEGER REFERENCES posts(id) ON DELETE SET NULL,
    comment_id   INTEGER REFERENCES comments(id) ON DELETE SET NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO files_backup (id, storage_path, original_name, mime_type, size, owner_user_id, post_id, comment_id, created_at)
    SELECT id, storage_path, original_name, mime_type, size, owner_user_id, post_id, comment_id, created_at FROM files;

DROP TABLE files;

ALTER TABLE files_backup RENAME TO files;

CREATE INDEX IF NOT EXISTS idx_files_owner ON files(owner_user_id);
