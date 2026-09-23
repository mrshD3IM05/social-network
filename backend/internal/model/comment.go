package model

import "time"

// NotificationCommentPost is sent to a post's author when someone comments on it.
const NotificationCommentPost = "comment_post"

type Comment struct {
	ID              int64     `json:"id"`
	PostID          int64     `json:"post_id"`
	AuthorID        int64     `json:"author_id"`
	AuthorFirstName string    `json:"author_first_name"`
	AuthorLastName  string    `json:"author_last_name"`
	AuthorNickname  string    `json:"author_nickname"`
	AuthorAvatar    string    `json:"author_avatar"`
	Content         string    `json:"content"`
	Images          []string  `json:"images"`
	CreatedAt       time.Time `json:"created_at"`
}
