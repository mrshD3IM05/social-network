package model

import "time"

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
	Likes           int       `json:"likes"`
	Dislikes        int       `json:"dislikes"`
	MyReaction      string    `json:"my_reaction"`
	CreatedAt       time.Time `json:"created_at"`
}
