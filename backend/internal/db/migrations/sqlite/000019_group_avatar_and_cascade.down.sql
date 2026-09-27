DROP TRIGGER IF EXISTS delete_comment_reactions;
DROP TRIGGER IF EXISTS delete_post_reactions;

CREATE TABLE notifications_old (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    type       TEXT NOT NULL,
    actor_id   INTEGER NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    group_id   INTEGER,
    read       INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
);

INSERT INTO notifications_old (id, user_id, type, actor_id, content, group_id, read, created_at)
    SELECT id, user_id, type, actor_id, content, group_id, read, created_at
    FROM notifications;

DROP TABLE notifications;

ALTER TABLE notifications_old RENAME TO notifications;

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, read);

ALTER TABLE groups DROP COLUMN avatar;
