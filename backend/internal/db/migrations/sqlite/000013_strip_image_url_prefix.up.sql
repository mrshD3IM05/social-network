-- Strip URL prefix from image columns, keeping only the UUID
UPDATE posts SET image = REPLACE(image, '/api/fs?f=', '') WHERE image LIKE '/api/fs?f=%';
UPDATE posts SET image = REPLACE(image, '/fileserve?f=', '') WHERE image LIKE '/fileserve?f=%';
UPDATE comments SET image = REPLACE(image, '/api/fs?f=', '') WHERE image LIKE '/api/fs?f=%';
UPDATE comments SET image = REPLACE(image, '/fileserve?f=', '') WHERE image LIKE '/fileserve?f=%';
