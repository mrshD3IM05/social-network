package repository

import "sn-backend/internal/model"

// the notification columns plus the name and photo of the user who caused it
const notificationSelect = `SELECT n.id, n.user_id, n.type, n.actor_id, u.first_name, u.last_name, u.avatar,
	n.content, n.group_id, n.read, n.created_at
	FROM notifications n JOIN users u ON u.id = n.actor_id`

func scanNotification(s scanner) (*model.Notification, error) {
	n := new(model.Notification)
	err := s.Scan(&n.ID, &n.UserID, &n.Type, &n.ActorID, &n.ActorFirstName, &n.ActorLastName, &n.ActorAvatar,
		&n.Content, &n.GroupID, &n.Read, &n.CreatedAt)
	return n, err
}

// CreateNotification stores the notification and fills in the rest of its
// fields (id, actor name, date) so it can be sent to the user as is.
func (r *Repository) CreateNotification(n *model.Notification) error {
	result, err := r.db.Exec(
		`INSERT INTO notifications (user_id, type, actor_id, content, group_id) VALUES (?, ?, ?, ?, ?)`,
		n.UserID, n.Type, n.ActorID, n.Content, n.GroupID,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	stored, err := scanNotification(r.QueryRow(notificationSelect+` WHERE n.id = ?`, id))
	if err != nil {
		return err
	}
	*n = *stored
	return nil
}

// ListNotifications returns the latest notifications of a user, newest first.
func (r *Repository) ListNotifications(userID int64) ([]*model.Notification, error) {
	rows, err := r.db.Query(notificationSelect+` WHERE n.user_id = ? ORDER BY n.id DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []*model.Notification{}
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *Repository) MarkNotificationsRead(userID int64) error {
	_, err := r.db.Exec(`UPDATE notifications SET read = 1 WHERE user_id = ? AND read = 0`, userID)
	return err
}
