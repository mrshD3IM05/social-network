CREATE TABLE IF NOT EXISTS reactions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    target_type TEXT NOT NULL CHECK (target_type IN ('post', 'comment')),
    target_id  INTEGER NOT NULL,
    user_id    INTEGER NOT NULL,
    reaction   TEXT NOT NULL CHECK (reaction IN ('like', 'dislike')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (target_type, target_id, user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reactions_target ON reactions(target_type, target_id);
