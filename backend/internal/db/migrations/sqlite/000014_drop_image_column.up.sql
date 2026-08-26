-- Drop legacy image column from posts (images now live in the files table)
ALTER TABLE posts DROP COLUMN image;

-- Drop legacy image column from comments
ALTER TABLE comments DROP COLUMN image;
