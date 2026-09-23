package repository

import "sn-backend/internal/model"

const commentColumns = `
	c.id, c.post_id, c.author_id, c.content, c.created_at,
	u.first_name, u.last_name, u.nickname, u.avatar`

func scanComment(s scanner) (*model.Comment, error) {
	comment := new(model.Comment)
	if err := s.Scan(
		&comment.ID,
		&comment.PostID,
		&comment.AuthorID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.AuthorFirstName,
		&comment.AuthorLastName,
		&comment.AuthorNickname,
		&comment.AuthorAvatar,
	); err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *Repository) CreateComment(comment *model.Comment) error {
	result, err := r.db.Exec(
		`INSERT INTO comments (post_id, author_id, content) VALUES (?, ?, ?)`,
		comment.PostID, comment.AuthorID, comment.Content,
	)
	if err != nil {
		return err
	}
	comment.ID, err = result.LastInsertId()
	return err
}

// GetComment loads one comment with its author fields.
func (r *Repository) GetComment(id int64) (*model.Comment, error) {
	comment, err := scanComment(r.QueryRow(
		`SELECT `+commentColumns+`
		 FROM comments c
		 JOIN users u ON u.id = c.author_id
		 WHERE c.id = ?`,
		id,
	))
	if err != nil {
		return nil, notFound(err)
	}
	return comment, nil
}

func (r *Repository) ListPostComments(postID int64) ([]*model.Comment, error) {
	rows, err := r.db.Query(
		`SELECT `+commentColumns+`
		 FROM comments c
		 JOIN users u ON u.id = c.author_id
		 WHERE c.post_id = ?
		 ORDER BY c.created_at, c.id`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*model.Comment, 0)
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comment.Images, err = r.ListCommentFileIDs(comment.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

// countPostComments returns the number of comments on a single post.
func (r *Repository) countPostComments(postID int64) (int, error) {
	var count int
	err := r.QueryRow(`SELECT COUNT(*) FROM comments WHERE post_id = ?`, postID).Scan(&count)
	return count, err
}

// CountPostComments returns the number of comments per post ID for the given
// posts, so post lists can show a count without N+1 queries.
func (r *Repository) CountPostComments(postIDs []int64) (map[int64]int, error) {
	counts := make(map[int64]int)
	if len(postIDs) == 0 {
		return counts, nil
	}
	rows, err := r.db.Query(
		`SELECT post_id, COUNT(*) FROM comments WHERE post_id IN (`+placeholders(len(postIDs))+`) GROUP BY post_id`,
		int64sToAny(postIDs)...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var postID int64
		var count int
		if err := rows.Scan(&postID, &count); err != nil {
			return nil, err
		}
		counts[postID] = count
	}
	return counts, rows.Err()
}

// CanViewComment mirrors CanViewPost: a comment is visible exactly when its
// post is visible to the viewer (post privacy for normal posts, group
// membership for group posts).
func (r *Repository) CanViewComment(viewerID, commentID int64) (bool, error) {
	var visible int
	err := r.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM comments c
			WHERE c.id = ? AND EXISTS (
				SELECT 1 FROM posts p WHERE p.id = c.post_id AND (`+postVisibleCondition+`)
			)
		)`,
		append([]any{commentID}, postVisibleArgs(viewerID)...)...,
	).Scan(&visible)
	return visible == 1, err
}

func (r *Repository) ListCommentFileIDs(commentID int64) ([]string, error) {
	rows, err := r.db.Query(`SELECT id FROM files WHERE comment_id = ? ORDER BY created_at, id`, commentID)
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
