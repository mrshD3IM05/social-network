package model

import "time"

const (
	PostPublic        = "public"
	PostFollowersOnly = "almost_private"
	PostSelected      = "private"
)

type Post struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	Privacy   string    `json:"privacy"`
	GroupID   *int64    `json:"group_id,omitempty"`
	Type      string    `json:"type"`
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"created_at"`
}
