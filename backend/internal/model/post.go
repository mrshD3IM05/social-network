package model

import "time"

const (
	PostPublic        = "public"
	PostFollowersOnly = "almost_private"
	PostSelected      = "private"
)

type Post struct {
	ID              int64     `json:"id"`
	AuthorID        int64     `json:"author_id"`
	AuthorFirstName string    `json:"author_first_name"`
	AuthorLastName  string    `json:"author_last_name"`
	AuthorNickname  string    `json:"author_nickname"`
	AuthorAvatar    string    `json:"author_avatar"`
	Content         string    `json:"content"`
	Privacy         string    `json:"privacy"`
	GroupID         *int64    `json:"group_id,omitempty"`
	Type            string    `json:"type"`
	Images          []string  `json:"images"`
	CreatedAt       time.Time `json:"created_at"`
}
