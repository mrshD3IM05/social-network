-- Strip legacy URL prefixes from users.avatar, keeping only the UUID
UPDATE users SET avatar = REPLACE(avatar, '/api/v1/fs?f=', '') WHERE avatar LIKE '/api/v1/fs?f=%';
UPDATE users SET avatar = REPLACE(avatar, '/api/fs?f=', '') WHERE avatar LIKE '/api/fs?f=%';
UPDATE users SET avatar = REPLACE(avatar, '/fileserve?f=', '') WHERE avatar LIKE '/fileserve?f=%';
