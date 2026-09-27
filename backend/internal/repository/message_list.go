package repository

import "sn-backend/internal/model"

// ListMessages returns the private messages between two users, oldest first.
// Only the last `limit` messages are kept, so a long conversation does not load
// all at once.
func (r *Repository) ListMessages(userID, otherID int64, limit int) ([]*model.Message, error) {
	// the inner query keeps the newest messages, the outer one puts them back
	// in reading order (same shape as ListGroupMessages)
	rows, err := r.db.Query(`
		SELECT id, from_user_id, to_user_id, group_id, content, created_at FROM (
			SELECT id, from_user_id, to_user_id, group_id, content, created_at
			FROM messages
			WHERE group_id IS NULL
			  AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))
			ORDER BY id DESC LIMIT ?
		) ORDER BY id`,
		userID, otherID, otherID, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*model.Message, 0)
	for rows.Next() {
		message := new(model.Message)
		if err := rows.Scan(
			&message.ID, &message.FromUserID, &message.ToUserID,
			&message.GroupID, &message.Content, &message.CreatedAt,
		); err != nil {
			return nil, err
		}
		message.Images, err = r.ListMessageFileIDs(message.ID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

// ListMessageFileIDs returns the ids of the images attached to one message.
func (r *Repository) ListMessageFileIDs(messageID int64) ([]string, error) {
	rows, err := r.db.Query(`SELECT id FROM files WHERE message_id = ? ORDER BY created_at, id`, messageID)
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
