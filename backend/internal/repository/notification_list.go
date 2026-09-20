package repository

import "sn-backend/internal/model"

// ListNotifications returns the newest notifications addressed to userID.
func (r *Repository) ListNotifications(userID int64, limit int) ([]*model.Notification, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, user_id, type, actor_id, content, group_id, read, created_at
		 FROM notifications WHERE user_id = ?
		 ORDER BY created_at DESC, id DESC LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*model.Notification, 0)
	for rows.Next() {
		n := new(model.Notification)
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.ActorID, &n.Content, &n.GroupID, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *Repository) CountUnreadNotifications(userID int64) (int, error) {
	var count int
	err := r.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read = 0`, userID).Scan(&count)
	return count, err
}

func (r *Repository) MarkNotificationRead(notificationID, userID int64) error {
	result, err := r.db.Exec(`UPDATE notifications SET read = 1 WHERE id = ? AND user_id = ?`, notificationID, userID)
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

func (r *Repository) MarkAllNotificationsRead(userID int64) error {
	_, err := r.db.Exec(`UPDATE notifications SET read = 1 WHERE user_id = ? AND read = 0`, userID)
	return err
}
