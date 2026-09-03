package model

import "time"

const (
	ReactionLike    = "like"
	ReactionDislike = "dislike"
)

const (
	ReactionTargetPost    = "post"
	ReactionTargetComment = "comment"
)

type Reaction struct {
	ID         int64     `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	UserID     int64     `json:"user_id"`
	Reaction   string    `json:"reaction"`
	CreatedAt  time.Time `json:"created_at"`
}

// ReactionSummary is what a viewer needs to render the like/dislike buttons:
// the totals plus their own reaction ("" when they have not reacted).
type ReactionSummary struct {
	Likes      int    `json:"likes"`
	Dislikes   int    `json:"dislikes"`
	MyReaction string `json:"my_reaction"`
}
