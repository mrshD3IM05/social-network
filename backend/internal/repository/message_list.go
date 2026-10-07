package repository

import "sn-backend/internal/model"

// ListMessages returns the private messages between two users, oldest first.
// Only the last `limit` messages are kept, so a long conversation does not load
// all at once.
func (r *MessageRepository) ListMessages(userID, otherID int64, limit int, lastID int64) ([]*model.Message, error) {
	// the inner query keeps the newest messages, the outer one puts them back
	// in reading order and reads the sender fields + images in one query.
	rows, err := r.db.Query(`
		SELECT `+messageViewColumns+`
		FROM (
			SELECT id FROM messages
			WHERE group_id IS NULL
			  AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))
			  AND (? = 0 OR id < ?)
			ORDER BY id DESC LIMIT ?
		) page
		JOIN message_view m ON m.id = page.id
		ORDER BY m.id`,
		userID, otherID, otherID, userID, lastID, lastID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*model.Message, 0)
	for rows.Next() {
		message, err := scanMessageView(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}
