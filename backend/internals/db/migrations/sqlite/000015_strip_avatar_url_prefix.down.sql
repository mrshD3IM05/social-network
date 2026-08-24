-- Restore URL prefix on users.avatar
UPDATE users SET avatar = '/api/fs?f=' || avatar WHERE avatar NOT LIKE '/%';
