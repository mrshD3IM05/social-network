package model

import "time"

const (
	FollowPending  = "pending"
	FollowAccepted = "accepted"
	FollowDeclined = "declined"
)

type FollowRequest struct {
	ID         int64     `json:"id"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   int64     `json:"to_user_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
