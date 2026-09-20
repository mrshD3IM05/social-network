package repository

import (
	"time"

	"sn-backend/internal/model"
)

// Conversation is one row of the inbox: the other participant plus the last
// message exchanged with them.
type Conversation struct {
	User        *model.User
	LastMessage *model.Message
}

func (r *Repository) scanMessages(query string, args ...any) ([]*model.Message, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]*model.Message, 0)
	for rows.Next() {
		message := new(model.Message)
		if err := rows.Scan(&message.ID, &message.FromUserID, &message.ToUserID, &message.GroupID, &message.Content, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// The query reads newest-first so LIMIT keeps the most recent page; the UI
	// wants them oldest-first.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

const messageColumns = `id, from_user_id, to_user_id, group_id, content, created_at`

// ListConversation returns the private messages exchanged between two users,
// newest page first. before is optional and pages backwards in time.
func (r *Repository) ListConversation(userID, otherID int64, limit int, before *time.Time) ([]*model.Message, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	if before != nil {
		return r.scanMessages(
			`SELECT `+messageColumns+` FROM messages
			 WHERE group_id IS NULL
			   AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))
			   AND created_at < ?
			 ORDER BY created_at DESC, id DESC LIMIT ?`,
			userID, otherID, otherID, userID, *before, limit,
		)
	}
	return r.scanMessages(
		`SELECT `+messageColumns+` FROM messages
		 WHERE group_id IS NULL
		   AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))
		 ORDER BY created_at DESC, id DESC LIMIT ?`,
		userID, otherID, otherID, userID, limit,
	)
}

// ListConversations returns every user the caller has already exchanged a
// private message with, most recent first.
func (r *Repository) ListConversations(userID int64) ([]*Conversation, error) {
	rows, err := r.db.Query(
		`SELECT `+followUserColumns+`,
		        m.id, m.from_user_id, m.to_user_id, m.group_id, m.content, m.created_at
		 FROM messages m
		 JOIN users u ON u.id = CASE WHEN m.from_user_id = ? THEN m.to_user_id ELSE m.from_user_id END
		 WHERE m.group_id IS NULL AND (m.from_user_id = ? OR m.to_user_id = ?)
		   AND m.id = (
			SELECT MAX(m2.id) FROM messages m2
			WHERE m2.group_id IS NULL
			  AND ((m2.from_user_id = m.from_user_id AND m2.to_user_id = m.to_user_id)
			    OR (m2.from_user_id = m.to_user_id AND m2.to_user_id = m.from_user_id))
		   )
		 ORDER BY m.created_at DESC, m.id DESC`,
		userID, userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]*Conversation, 0)
	for rows.Next() {
		user := new(model.User)
		message := new(model.Message)
		var private int
		if err := rows.Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Nickname, &user.Avatar, &user.AboutMe, &private, &user.CreatedAt,
			&message.ID, &message.FromUserID, &message.ToUserID, &message.GroupID, &message.Content, &message.CreatedAt,
		); err != nil {
			return nil, err
		}
		user.Private = private == 1
		conversations = append(conversations, &Conversation{User: user, LastMessage: message})
	}
	return conversations, rows.Err()
}
