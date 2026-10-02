-- user_view is a user plus the two follow counts, so every read that shows a
-- user does not have to repeat them. Only an accepted request counts.
--
-- is_followed and is_following are deliberately left out: they are relative to
-- whoever is asking and a view cannot be handed an id, so they stay on the
-- queries that know the viewer.
CREATE VIEW user_view AS
SELECT
    u.id,
    u.email,
    u.password,
    u.first_name,
    u.last_name,
    u.date_of_birth,
    u.avatar,
    u.nickname,
    u.about_me,
    u.private,
    u.created_at,
    (SELECT COUNT(*) FROM follow_requests f
      WHERE f.to_user_id = u.id AND f.status = 'accepted') AS followers,
    (SELECT COUNT(*) FROM follow_requests f
      WHERE f.from_user_id = u.id AND f.status = 'accepted') AS following,
    (SELECT COUNT(*) FROM posts p
      WHERE p.author_id = u.id AND p.group_id IS NULL) AS post_count
FROM users u;
