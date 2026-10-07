package repository

import (
	"encoding/json"

	"sn-backend/internal/model"
)

const messageViewColumns = `
	m.id, m.from_user_id, m.to_user_id, m.group_id, m.content, m.created_at,
	m.first_name, m.last_name, m.avatar, m.images`

// scanMessageView reads a message_view row (migration 000024): the message, the
// sender's display fields and the images JSON array.
func scanMessageView(s scanner) (*model.Message, error) {
	message := new(model.Message)
	var images string
	if err := s.Scan(
		&message.ID,
		&message.FromUserID,
		&message.ToUserID,
		&message.GroupID,
		&message.Content,
		&message.CreatedAt,
		&message.FromFirstName,
		&message.FromLastName,
		&message.FromAvatar,
		&images,
	); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(images), &message.Images); err != nil {
		return nil, err
	}
	return message, nil
}

func (r *MessageRepository) CreateMessage(message *model.Message) error {
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
	return r.QueryRow(
		`SELECT m.created_at, m.first_name, m.last_name, m.avatar
		 FROM message_view m WHERE m.id = ?`, message.ID,
	).Scan(&message.CreatedAt, &message.FromFirstName, &message.FromLastName, &message.FromAvatar)
}

func (r *MessageRepository) CanMessage(fromUserID int64, toUserID, groupID *int64) (bool, error) {
	if toUserID != nil {
		var allowed int
		// At least one of the two must follow the other.
		// A public profile is not enough on its own.
		err := r.QueryRow(`SELECT EXISTS(
			SELECT 1 FROM users target
			WHERE target.id = ? AND EXISTS (
				SELECT 1 FROM follow_requests f
				WHERE (f.from_user_id = ? AND f.to_user_id = target.id OR f.from_user_id = target.id AND f.to_user_id = ?)
				AND f.status = 'accepted'
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

// ListGroupMessages returns one older-to-newer page of a group chat.
func (r *MessageRepository) ListGroupMessages(groupID, lastID int64) ([]*model.Message, error) {
	// the inner query keeps the newest messages, the outer one puts them back
	// in reading order and reads the sender fields + images in one query.
	rows, err := r.db.Query(`
		SELECT `+messageViewColumns+`
		FROM (
			SELECT id FROM messages
			WHERE group_id = ? AND (? = 0 OR id < ?)
			ORDER BY id DESC LIMIT ?
		) page
		JOIN message_view m ON m.id = page.id
		ORDER BY m.id`,
		groupID, lastID, lastID, MessagePageSize,
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
