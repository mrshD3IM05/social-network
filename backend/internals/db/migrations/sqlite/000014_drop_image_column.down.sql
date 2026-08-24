-- Recreate tables with image column (best-effort rollback)
CREATE TABLE IF NOT EXISTS posts_new (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    author_id  INTEGER NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    image      TEXT NOT NULL DEFAULT '',
    privacy    TEXT NOT NULL DEFAULT 'public',
    group_id   INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO posts_new (id, author_id, content, privacy, group_id, created_at)
    SELECT id, author_id, content, privacy, group_id, created_at FROM posts;
DROP TABLE posts;
ALTER TABLE posts_new RENAME TO posts;

CREATE TABLE IF NOT EXISTS comments_new (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id    INTEGER NOT NULL,
    author_id  INTEGER NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    image      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO comments_new (id, post_id, author_id, content, created_at)
    SELECT id, post_id, author_id, content, created_at FROM comments;
DROP TABLE comments;
ALTER TABLE comments_new RENAME TO comments;
