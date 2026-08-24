CREATE TABLE IF NOT EXISTS follow_requests (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    from_user_id INTEGER NOT NULL,
    to_user_id   INTEGER NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (from_user_id, to_user_id)
);

CREATE INDEX IF NOT EXISTS idx_follow_requests_to ON follow_requests(to_user_id, status);
CREATE INDEX IF NOT EXISTS idx_follow_requests_from ON follow_requests(from_user_id, status);
