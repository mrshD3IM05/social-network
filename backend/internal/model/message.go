package model

import "time"

type Message struct {
	ID         int64     `json:"id"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   *int64    `json:"to_user_id,omitempty"`
	GroupID    *int64    `json:"group_id,omitempty"`
	Content    string    `json:"content"`
	Images     []string  `json:"images"`
	CreatedAt  time.Time `json:"created_at"`
	// the sender, so a group chat can name who wrote each message without
	// loading every member of the group
	FromFirstName string `json:"from_first_name"`
	FromLastName  string `json:"from_last_name"`
	FromAvatar    string `json:"from_avatar"`
}

const (
	NotificationFollowRequest  = "follow_request"
	NotificationNewFollower    = "new_follower"
	NotificationFollowAccepted = "follow_accepted"
)

type Notification struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"-"` // the recipient, who is always the viewer
	Type           string    `json:"type"`
	ActorID        int64     `json:"actor_id"`
	ActorFirstName string    `json:"actor_first_name"`
	ActorLastName  string    `json:"actor_last_name"`
	ActorAvatar    string    `json:"actor_avatar"`
	Content        string    `json:"content"`
	GroupID        *int64    `json:"group_id,omitempty"`
	Read           bool      `json:"read"`
	CreatedAt      time.Time `json:"created_at"`
}
