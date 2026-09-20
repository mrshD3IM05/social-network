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

func (r *Repository) GetComment(commentID int64) (*model.Comment, error) {
	comment, err := scanComment(r.QueryRow(
		`SELECT `+commentColumns+`
		FROM comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.id = ?`,
		commentID,
	))
	if err != nil {
		return nil, notFound(err)
	}
	comment.Images, err = r.ListCommentFileIDs(comment.ID)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *Repository) ListPostComments(postID, viewerID int64) ([]*model.Comment, error) {
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
		summary, err := r.GetReactionSummary(model.ReactionTargetComment, comment.ID, viewerID)
		if err != nil {
			return nil, err
		}
		comment.Likes, comment.Dislikes, comment.MyReaction = summary.Likes, summary.Dislikes, summary.MyReaction
		comments = append(comments, comment)
	}
	return comments, rows.Err()
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

// AttachFileToComment binds an already stored upload to a comment the uploader
// owns. Comment images also carry the comment's post_id so that the existing
// per-viewer file visibility query resolves them through the post.
func (r *Repository) AttachFileToComment(fileID string, commentID, ownerID int64) error {
	result, err := r.db.Exec(
		`UPDATE files
		 SET comment_id = ?, post_id = (SELECT post_id FROM comments WHERE id = ?)
		 WHERE id = ? AND owner_user_id = ?`,
		commentID, commentID, fileID, ownerID,
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

// DeleteCommentOwned removes a comment the caller wrote, or any comment on a
// post the caller owns, together with the file rows attached to it.
func (r *Repository) DeleteCommentOwned(commentID, ownerID int64) ([]string, error) {
	fileIDs, err := r.ListCommentFileIDs(commentID)
	if err != nil {
		return nil, err
	}
	result, err := r.db.Exec(
		`DELETE FROM comments
		 WHERE id = ? AND (
			author_id = ? OR EXISTS (SELECT 1 FROM posts p WHERE p.id = comments.post_id AND p.author_id = ?)
		 )`,
		commentID, ownerID, ownerID,
	)
	if err != nil {
		return nil, err
	}
	if count, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if count == 0 {
		return nil, ErrNotFound
	}
	if _, err := r.db.Exec(`DELETE FROM files WHERE comment_id = ?`, commentID); err != nil {
		return nil, err
	}
	if _, err := r.db.Exec(
		`DELETE FROM reactions WHERE target_type = ? AND target_id = ?`,
		model.ReactionTargetComment, commentID,
	); err != nil {
		return nil, err
	}
	return fileIDs, nil
}

func (r *Repository) CommentPostID(commentID int64) (int64, error) {
	var postID int64
	err := r.QueryRow(`SELECT post_id FROM comments WHERE id = ?`, commentID).Scan(&postID)
	if err != nil {
		return 0, notFound(err)
	}
	return postID, nil
}
