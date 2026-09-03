package repository

import "sn-backend/internal/model"

func (r *Repository) CreatePost(post *model.Post) error {
	result, err := r.db.Exec(
		`INSERT INTO posts (author_id, content, privacy, group_id, type) VALUES (?, ?, ?, ?, ?)`,
		post.AuthorID, post.Content, post.Privacy, post.GroupID, post.Type,
	)
	if err != nil {
		return err
	}
	post.ID, err = result.LastInsertId()
	return err
}

const postColumns = `
	p.id, p.author_id, p.content, p.privacy, p.group_id, p.type, p.created_at,
	u.first_name, u.last_name, u.nickname, u.avatar`

func scanPost(s scanner) (*model.Post, error) {
	post := new(model.Post)
	if err := s.Scan(
		&post.ID,
		&post.AuthorID,
		&post.Content,
		&post.Privacy,
		&post.GroupID,
		&post.Type,
		&post.CreatedAt,
		&post.AuthorFirstName,
		&post.AuthorLastName,
		&post.AuthorNickname,
		&post.AuthorAvatar,
	); err != nil {
		return nil, err
	}
	return post, nil
}

func (r *Repository) GetPost(postID int64) (*model.Post, error) {
	post, err := scanPost(r.QueryRow(
		`SELECT `+postColumns+`
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id = ?`,
		postID,
	))
	if err != nil {
		return nil, notFound(err)
	}
	post.Images, err = r.ListPostFileIDs(post.ID)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *Repository) ListPostFileIDs(postID int64) ([]string, error) {
	rows, err := r.db.Query(`SELECT id FROM files WHERE post_id = ? ORDER BY created_at, id`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) UpdatePostOwned(post *model.Post, ownerID int64) error {
	result, err := r.db.Exec(
		`UPDATE posts SET content = ?, privacy = ? WHERE id = ? AND author_id = ?`,
		post.Content, post.Privacy, post.ID, ownerID,
	)
	if err != nil {
		return err
	}
	if count, err := result.RowsAffected(); err != nil {
		return err
	} else if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeletePostOwned(postID, ownerID int64) error {
	result, err := r.db.Exec(`DELETE FROM posts WHERE id = ? AND author_id = ?`, postID, ownerID)
	if err != nil {
		return err
	}
	if count, err := result.RowsAffected(); err != nil {
		return err
	} else if count == 0 {
		return ErrNotFound
	}
	return nil
}

// postVisibleCondition is the "may this viewer see post p" predicate, shared by
// the feed and by the single-post check. It takes the arguments returned by
// postVisibleArgs, in that order.
const postVisibleCondition = `(
	p.author_id = ? OR p.privacy = ? OR
	(p.privacy = ? AND EXISTS (
		SELECT 1 FROM follow_requests f
		WHERE f.from_user_id = ? AND f.to_user_id = p.author_id AND f.status = ?
	)) OR
	(p.privacy = ? AND EXISTS (
		SELECT 1 FROM post_visibility v
		WHERE v.post_id = p.id AND v.user_id = ?
	))
)`

func postVisibleArgs(viewerID int64) []any {
	return []any{
		viewerID, model.PostPublic, model.PostFollowersOnly, viewerID, model.FollowAccepted,
		model.PostSelected, viewerID,
	}
}

func (r *Repository) ListVisiblePosts(viewerID int64) ([]*model.Post, error) {
	rows, err := r.db.Query(`
		SELECT `+postColumns+`
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.group_id IS NULL AND `+postVisibleCondition+`
		ORDER BY p.created_at DESC, p.id DESC`,
		postVisibleArgs(viewerID)...,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// CanViewPost reports whether viewerID is allowed to see the post, using the
// same rules as the feed.
func (r *Repository) CanViewPost(viewerID, postID int64) (bool, error) {
	var visible bool
	args := append([]any{postID}, postVisibleArgs(viewerID)...)
	err := r.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM posts p WHERE p.id = ? AND `+postVisibleCondition+`)`,
		args...,
	).Scan(&visible)
	return visible, err
}

// LoadPostReactions fills in the post's like/dislike totals and the viewer's
// own reaction.
func (r *Repository) LoadPostReactions(post *model.Post, viewerID int64) error {
	summary, err := r.GetReactionSummary(model.ReactionTargetPost, post.ID, viewerID)
	if err != nil {
		return err
	}
	post.Likes = summary.Likes
	post.Dislikes = summary.Dislikes
	post.MyReaction = summary.MyReaction
	return nil
}
