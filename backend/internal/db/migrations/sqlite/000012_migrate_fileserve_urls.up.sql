-- Migrate old /fileserve URLs to /api/fs
UPDATE posts SET image = REPLACE(image, '/fileserve?f=', '/api/fs?f=') WHERE image LIKE '%fileserve%';
UPDATE comments SET image = REPLACE(image, '/fileserve?f=', '/api/fs?f=') WHERE image LIKE '%fileserve%';
