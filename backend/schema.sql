-- =============================================================================
-- Social network backend - consolidated SQLite schema (DDL)
-- Generated from internals/db/migrations/sqlite (000001..000017).
-- Migrations are embedded and applied automatically at startup; this file is a
-- reference snapshot of the resulting schema.
-- =============================================================================

PRAGMA foreign_keys = ON;

-- -----------------------------------------------------------------------------
-- Identity & sessions
-- -----------------------------------------------------------------------------

CREATE TABLE users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    password      TEXT NOT NULL,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    date_of_birth TEXT NOT NULL,
    avatar        TEXT NOT NULL DEFAULT '',
    nickname      TEXT NOT NULL DEFAULT '',
    about_me      TEXT NOT NULL DEFAULT '',
    private       INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id         TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE follow_requests (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    from_user_id INTEGER NOT NULL,
    to_user_id   INTEGER NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (from_user_id, to_user_id)
);

CREATE INDEX idx_follow_requests_to ON follow_requests(to_user_id, status);
CREATE INDEX idx_follow_requests_from ON follow_requests(from_user_id, status);

-- -----------------------------------------------------------------------------
-- Posts & comments
-- -----------------------------------------------------------------------------

CREATE TABLE posts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    author_id  INTEGER NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    privacy    TEXT NOT NULL DEFAULT 'public',
    group_id   INTEGER,
    type       TEXT NOT NULL DEFAULT 'post',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_group ON posts(group_id);

CREATE TABLE post_visibility (
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE comments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id    INTEGER NOT NULL,
    author_id  INTEGER NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_comments_post ON comments(post_id);

CREATE TABLE reactions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    target_type TEXT NOT NULL CHECK (target_type IN ('post', 'comment')),
    target_id   INTEGER NOT NULL,
    user_id     INTEGER NOT NULL,
    reaction    TEXT NOT NULL CHECK (reaction IN ('like', 'dislike')),
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (target_type, target_id, user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_reactions_target ON reactions(target_type, target_id);

-- -----------------------------------------------------------------------------
-- Groups & events
-- -----------------------------------------------------------------------------

CREATE TABLE groups (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    creator_id  INTEGER NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_members (
    group_id   INTEGER NOT NULL,
    user_id    INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_invitations (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id     INTEGER NOT NULL,
    from_user_id INTEGER NOT NULL,
    to_user_id   INTEGER NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (group_id, to_user_id)
);

CREATE INDEX idx_group_invites_to ON group_invitations(to_user_id, status);

CREATE TABLE group_join_requests (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id   INTEGER NOT NULL,
    user_id    INTEGER NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (group_id, user_id)
);

CREATE INDEX idx_group_join_requests_group ON group_join_requests(group_id, status);

CREATE TABLE group_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id    INTEGER NOT NULL,
    creator_id  INTEGER NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    date_time   TIMESTAMP NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_group_events_group ON group_events(group_id);

CREATE TABLE event_responses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id   INTEGER NOT NULL,
    user_id    INTEGER NOT NULL,
    choice     TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (event_id, user_id)
);

-- -----------------------------------------------------------------------------
-- Notifications & messages
-- -----------------------------------------------------------------------------

CREATE TABLE notifications (
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

CREATE INDEX idx_notifications_user ON notifications(user_id, read);

-- rebuilt by migration 000008 and again by 000017 (final shape below):
-- the CHECK allows group messages where to_user_id IS NULL
CREATE TABLE messages (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    from_user_id INTEGER NOT NULL,
    to_user_id   INTEGER,
    group_id     INTEGER,
    content      TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CHECK (from_user_id <> to_user_id OR to_user_id IS NULL)
);

CREATE INDEX idx_messages_to ON messages(to_user_id);
CREATE INDEX idx_messages_group ON messages(group_id);

-- -----------------------------------------------------------------------------
-- Stored files (originals on disk under uploads/, thumbnails derived from path)
-- -----------------------------------------------------------------------------

CREATE TABLE files (
    id            TEXT PRIMARY KEY,
    storage_path  TEXT NOT NULL UNIQUE,
    original_name TEXT NOT NULL,
    mime_type     TEXT NOT NULL,
    size          INTEGER NOT NULL,
    owner_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id       INTEGER REFERENCES posts(id) ON DELETE SET NULL,
    comment_id    INTEGER REFERENCES comments(id) ON DELETE SET NULL,
    message_id    INTEGER REFERENCES messages(id) ON DELETE SET NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_files_owner ON files(owner_user_id);
CREATE INDEX idx_files_message ON files(message_id);

-- -----------------------------------------------------------------------------
-- Migration bookkeeping (managed by golang-migrate, not part of the app schema)
-- -----------------------------------------------------------------------------

CREATE TABLE schema_migrations (version uint64, dirty bool);
CREATE UNIQUE INDEX version_unique ON schema_migrations (version);
