CREATE TABLE IF NOT EXISTS users (
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
