-- Reverse: restore old /fileserve URLs
UPDATE posts SET image = REPLACE(image, '/api/fs?f=', '/fileserve?f=') WHERE image LIKE '%/api/fs%';
UPDATE comments SET image = REPLACE(image, '/api/fs?f=', '/fileserve?f=') WHERE image LIKE '%/api/fs%';
