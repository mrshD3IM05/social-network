-- Login accepts a nickname as an identifier, so duplicates make an account
-- ambiguous. Existing collisions are renamed before the index is created.
UPDATE users
SET nickname = substr(nickname, 1, 10) || id
WHERE nickname <> ''
  AND id NOT IN (SELECT MIN(id) FROM users WHERE nickname <> '' GROUP BY nickname);

-- The index is partial: the nickname is optional, so '' may repeat freely.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_nickname_unique
    ON users(nickname) WHERE nickname <> '';
