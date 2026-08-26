-- Restore URL prefix
UPDATE posts SET image = '/api/fs?f=' || image WHERE image NOT LIKE '/%';
UPDATE comments SET image = '/api/fs?f=' || image WHERE image NOT LIKE '/%';
