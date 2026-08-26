CREATE TABLE IF NOT EXISTS files (
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

CREATE INDEX IF NOT EXISTS idx_files_owner ON files(owner_user_id);
