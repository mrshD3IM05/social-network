CREATE TABLE messages_old (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    from_user_id INTEGER NOT NULL,
    to_user_id   INTEGER,
    group_id     INTEGER,
    content      TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

INSERT INTO messages_old (id, from_user_id, to_user_id, group_id, content, created_at)
    SELECT id, from_user_id, to_user_id, group_id, content, created_at
    FROM messages;

DROP TABLE messages;

ALTER TABLE messages_old RENAME TO messages;

CREATE INDEX IF NOT EXISTS idx_messages_to ON messages(to_user_id);
CREATE INDEX IF NOT EXISTS idx_messages_group ON messages(group_id);
