package repository

import "sn-backend/internal/model"

func (r *Repository) CreateMessage(message *model.Message) error {
	result, err := r.db.Exec(`
		INSERT INTO messages (from_user_id, to_user_id, group_id, content)
		VALUES (?, ?, ?, ?)`, message.FromUserID, message.ToUserID, message.GroupID, message.Content)
	if err != nil {
		return err
	}
	message.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return r.QueryRow(`SELECT created_at FROM messages WHERE id = ?`, message.ID).Scan(&message.CreatedAt)
}

func (r *Repository) CanMessage(fromUserID int64, toUserID, groupID *int64) (bool, error) {
	if toUserID != nil {
		var allowed int
		err := r.QueryRow(`SELECT EXISTS(
			SELECT 1 FROM users target
			WHERE target.id = ? AND (
				 target.private = 0 OR EXISTS (
					SELECT 1 FROM follow_requests f
					WHERE (f.from_user_id = ? AND f.to_user_id = target.id OR f.from_user_id = target.id AND f.to_user_id = ?)
					AND f.status = 'accepted'
				)
			)
		)`, *toUserID, fromUserID, fromUserID).Scan(&allowed)
		return allowed == 1, err
	}
	if groupID != nil {
		var allowed int
		err := r.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)`, *groupID, fromUserID).Scan(&allowed)
		return allowed == 1, err
	}
	return false, nil
}

func (r *Repository) GroupMemberIDs(groupID int64) ([]int64, error) {
	rows, err := r.db.Query(`SELECT user_id FROM group_members WHERE group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		members = append(members, userID)
	}
	return members, rows.Err()
}
