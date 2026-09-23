package repository

import "sn-backend/internal/model"

func (r *Repository) CreateFile(file *model.File) error {
	_, err := r.db.Exec(`
		INSERT INTO files (id, storage_path, original_name, mime_type, size, owner_user_id, post_id, comment_id, message_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		file.ID, file.StoragePath, file.OriginalName, file.MIMEType, file.Size,
		file.OwnerUserID, file.PostID, file.CommentID, file.MessageID,
	)
	return err
}

func (r *Repository) GetFile(id string) (*model.File, error) {
	file := new(model.File)
	err := r.QueryRow(`
		SELECT id, storage_path, original_name, mime_type, size, owner_user_id, post_id, comment_id, message_id, created_at
		FROM files WHERE id = ?`, id,
	).Scan(
		&file.ID, &file.StoragePath, &file.OriginalName, &file.MIMEType, &file.Size,
		&file.OwnerUserID, &file.PostID, &file.CommentID, &file.MessageID, &file.CreatedAt,
	)
	if err != nil {
		return nil, notFound(err)
	}
	return file, nil
}

func (r *Repository) CanViewFile(viewerID int64, fileID string) (bool, error) {
	var visible int
	err := r.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM files f
			LEFT JOIN posts p ON p.id = f.post_id
			LEFT JOIN messages m ON m.id = f.message_id
			WHERE f.id = ? AND (
				f.owner_user_id = ? OR
				EXISTS (SELECT 1 FROM users u WHERE u.avatar = f.id) OR
				EXISTS (SELECT 1 FROM posts p2 WHERE p2.id = f.post_id AND p2.group_id IS NOT NULL AND EXISTS (
					SELECT 1 FROM group_members gmp
					WHERE gmp.group_id = p2.group_id AND gmp.user_id = ?
				)) OR
				EXISTS (SELECT 1 FROM comments cm WHERE cm.id = f.comment_id AND EXISTS (
					SELECT 1 FROM posts p3 WHERE p3.id = cm.post_id AND (
						(p3.group_id IS NOT NULL AND EXISTS (
							SELECT 1 FROM group_members gmc
							WHERE gmc.group_id = p3.group_id AND gmc.user_id = ?
						))
						OR
						(p3.group_id IS NULL AND (
							p3.author_id = ? OR p3.privacy = ? OR
							(p3.privacy = ? AND EXISTS (
								SELECT 1 FROM follow_requests fr2
								WHERE fr2.from_user_id = ? AND fr2.to_user_id = p3.author_id AND fr2.status = ?
							)) OR
							(p3.privacy = ? AND EXISTS (
								SELECT 1 FROM post_visibility pv2
								WHERE pv2.post_id = p3.id AND pv2.user_id = ?
							))
						))
					)
				)) OR
				p.author_id = ? OR p.privacy = ? OR
				(p.privacy = ? AND EXISTS (
					SELECT 1 FROM follow_requests fr
					WHERE fr.from_user_id = ? AND fr.to_user_id = p.author_id AND fr.status = ?
				)) OR
				(p.privacy = ? AND EXISTS (
					SELECT 1 FROM post_visibility pv
					WHERE pv.post_id = p.id AND pv.user_id = ?
				)) OR
				(m.from_user_id = ? OR m.to_user_id = ? OR (m.group_id IS NOT NULL AND EXISTS (
					SELECT 1 FROM group_members gm
					WHERE gm.group_id = m.group_id AND gm.user_id = ?
				)))
			)
		)`,
		fileID, viewerID, viewerID, viewerID, viewerID, model.PostPublic,
		model.PostFollowersOnly, viewerID, model.FollowAccepted, model.PostSelected, viewerID,
		viewerID, model.PostPublic, model.PostFollowersOnly, viewerID, model.FollowAccepted,
		model.PostSelected, viewerID, viewerID, viewerID, viewerID,
	).Scan(&visible)
	return visible == 1, err
}
