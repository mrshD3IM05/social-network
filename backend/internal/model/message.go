package model

import "time"

type Message struct {
	ID         int64     `json:"id"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   *int64    `json:"to_user_id,omitempty"`
	GroupID    *int64    `json:"group_id,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	ActorID   int64     `json:"actor_id"`
	Content   string    `json:"content"`
	GroupID   *int64    `json:"group_id,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}
