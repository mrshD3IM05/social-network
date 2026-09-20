package repository

import "sn-backend/internal/model"

// SetPostAudience replaces the chosen-follower list of a "private" post.
// Only accepted followers of the author are kept, so a caller cannot widen a
// post to users who do not follow them.
func (r *Repository) SetPostAudience(postID, authorID int64, userIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var owner int64
	if err := tx.QueryRow(`SELECT author_id FROM posts WHERE id = ?`, postID).Scan(&owner); err != nil {
		return notFound(err)
	}
	if owner != authorID {
		return ErrNotOwner
	}
	if _, err := tx.Exec(`DELETE FROM post_visibility WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if userID == authorID {
			continue
		}
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO post_visibility (post_id, user_id)
			 SELECT ?, ? WHERE EXISTS (
				SELECT 1 FROM follow_requests
				WHERE from_user_id = ? AND to_user_id = ? AND status = ?
			 )`,
			postID, userID, userID, authorID, model.FollowAccepted,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) ListPostAudience(postID int64) ([]int64, error) {
	rows, err := r.db.Query(`SELECT user_id FROM post_visibility WHERE post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListUserPosts returns the posts written by authorID that viewerID is allowed
// to see, so a profile page no longer has to filter the whole feed.
func (r *Repository) ListUserPosts(authorID, viewerID int64) ([]*model.Post, error) {
	args := append([]any{authorID}, postVisibleArgs(viewerID)...)
	rows, err := r.db.Query(`
		SELECT `+postColumns+`
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.group_id IS NULL AND p.author_id = ? AND `+postVisibleCondition+`
		ORDER BY p.created_at DESC, p.id DESC`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*model.Post, 0)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		post.Images, err = r.ListPostFileIDs(post.ID)
		if err != nil {
			return nil, err
		}
		if err := r.LoadPostReactions(post, viewerID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}
