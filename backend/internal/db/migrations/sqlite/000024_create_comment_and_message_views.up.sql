-- comment_view is a comment plus its author's display fields and the ids of its
-- attached images as a JSON array, so listing a post's comments needs one query
-- instead of one extra file lookup per comment (the same shape as post_view).
CREATE VIEW comment_view AS
SELECT
    c.id,
    c.post_id,
    c.author_id,
    c.content,
    c.created_at,
    u.first_name,
    u.last_name,
    u.nickname,
    u.avatar,
    COALESCE(
        (SELECT json_group_array(f.id ORDER BY f.created_at, f.id)
         FROM files f
         WHERE f.comment_id = c.id),
        '[]'
    ) AS images
FROM comments c
JOIN users u ON u.id = c.author_id;

-- message_view: a message plus the sender's display fields and its image ids as
-- JSON, so listing a conversation needs one query instead of one per message.
CREATE VIEW message_view AS
SELECT
    m.id,
    m.from_user_id,
    m.to_user_id,
    m.group_id,
    m.content,
    m.created_at,
    u.first_name,
    u.last_name,
    COALESCE(u.avatar, '') AS avatar,
    COALESCE(
        (SELECT json_group_array(f.id ORDER BY f.created_at, f.id)
         FROM files f
         WHERE f.message_id = m.id),
        '[]'
    ) AS images
FROM messages m
JOIN users u ON u.id = m.from_user_id;

CREATE INDEX IF NOT EXISTS idx_files_comment ON files(comment_id);
CREATE INDEX IF NOT EXISTS idx_files_post ON files(post_id);