-- Group profile picture (a file id, like users.avatar)
ALTER TABLE groups ADD COLUMN avatar TEXT NOT NULL DEFAULT '';

-- Group notifications are deleted with their group instead of being kept
-- with a NULL group_id
CREATE TABLE notifications_new (
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
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

INSERT INTO notifications_new (id, user_id, type, actor_id, content, group_id, read, created_at)
    SELECT id, user_id, type, actor_id, content, group_id, read, created_at
    FROM notifications;

DROP TABLE notifications;

ALTER TABLE notifications_new RENAME TO notifications;

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, read);

-- reactions has no foreign key (target_id points to posts or comments), so
-- triggers remove them when the post or comment goes away (this also covers
-- the cascade group -> posts -> comments)
DELETE FROM reactions WHERE target_type = 'post' AND target_id NOT IN (SELECT id FROM posts);
DELETE FROM reactions WHERE target_type = 'comment' AND target_id NOT IN (SELECT id FROM comments);

CREATE TRIGGER IF NOT EXISTS delete_post_reactions AFTER DELETE ON posts
BEGIN
    DELETE FROM reactions WHERE target_type = 'post' AND target_id = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS delete_comment_reactions AFTER DELETE ON comments
BEGIN
    DELETE FROM reactions WHERE target_type = 'comment' AND target_id = OLD.id;
END;
