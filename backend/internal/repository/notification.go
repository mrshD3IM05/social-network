package repository

import (
	"database/sql"

	"sn-backend/internal/model"
)

// CreateNotification inserts a notification row. Used by services that must
// notify a user about group activity; hub.PublishNotification fans it out to
// the user's connected websockets.
func (r *Repository) CreateNotification(n *model.Notification) error {
	result, err := r.db.Exec(
		`INSERT INTO notifications (user_id, type, actor_id, content, group_id) VALUES (?, ?, ?, ?, ?)`,
		n.UserID, n.Type, n.ActorID, n.Content, n.GroupID,
	)
	if err != nil {
		return err
	}
	n.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return r.QueryRow(`SELECT read, created_at FROM notifications WHERE id = ?`, n.ID).Scan(&n.Read, &n.CreatedAt)
}

// CreateNotificationTx is the transaction-aware variant used inside the
// accept/refuse flows so the notification and the state change commit together.
func (r *Repository) CreateNotificationTx(n *model.Notification, tx *sql.Tx) error {
	result, err := tx.Exec(
		`INSERT INTO notifications (user_id, type, actor_id, content, group_id) VALUES (?, ?, ?, ?, ?)`,
		n.UserID, n.Type, n.ActorID, n.Content, n.GroupID,
	)
	if err != nil {
		return err
	}
	n.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return tx.QueryRow(`SELECT read, created_at FROM notifications WHERE id = ?`, n.ID).Scan(&n.Read, &n.CreatedAt)
}
