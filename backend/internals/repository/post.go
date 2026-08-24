package repository

import "sn-backend/internals/model"

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

func (r *Repository) GetPost(postID int64) (*model.Post, error) {
	post := new(model.Post)
	err := r.QueryRow(
		`SELECT id, author_id, content, privacy, group_id, type, created_at FROM posts WHERE id = ?`,
		postID,
	).Scan(&post.ID, &post.AuthorID, &post.Content, &post.Privacy, &post.GroupID, &post.Type, &post.CreatedAt)
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

func (r *Repository) ListVisiblePosts(viewerID int64) ([]*model.Post, error) {
	rows, err := r.db.Query(`
		SELECT id, author_id, content, privacy, group_id, type, created_at
		FROM posts
		WHERE group_id IS NULL AND (
			author_id = ? OR privacy = ? OR
			(privacy = ? AND EXISTS (
				SELECT 1 FROM follow_requests f
				WHERE f.from_user_id = ? AND f.to_user_id = posts.author_id AND f.status = ?
			)) OR
			(privacy = ? AND EXISTS (
				SELECT 1 FROM post_visibility v
				WHERE v.post_id = posts.id AND v.user_id = ?
			))
		)
		ORDER BY created_at DESC, id DESC`,
		viewerID, model.PostPublic, model.PostFollowersOnly, viewerID, model.FollowAccepted,
		model.PostSelected, viewerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*model.Post, 0)
	for rows.Next() {
		post := new(model.Post)
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Content, &post.Privacy, &post.GroupID, &post.Type, &post.CreatedAt); err != nil {
			return nil, err
		}
		post.Images, err = r.ListPostFileIDs(post.ID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
