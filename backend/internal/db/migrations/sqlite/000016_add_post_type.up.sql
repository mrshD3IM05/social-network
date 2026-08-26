-- Add type column to posts (e.g. 'post', 'avatar_update')
ALTER TABLE posts ADD COLUMN type TEXT NOT NULL DEFAULT 'post';
