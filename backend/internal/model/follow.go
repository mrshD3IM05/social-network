package model

import "time"

const (
	FollowPending  = "pending"
	FollowAccepted = "accepted"
	FollowDeclined = "declined"
)

// The numbers User.IsFollowed and User.IsFollowing carry, so a client reads
// the relation of two users without parsing a status word.
const (
	FollowStateNone    = 0
	FollowStateActive  = 1
	FollowStatePending = 2
)

// FollowState turns a follow_requests.status into the number User carries:
// accepted is an active follow, pending is a request waiting for an answer,
// and a declined request reads as no relation at all — the same reading
// followsvc.Status gives it.
func FollowState(status string) int64 {
	switch status {
	case FollowAccepted:
		return FollowStateActive
	case FollowPending:
		return FollowStatePending
	default:
		return FollowStateNone
	}
}

type FollowRequest struct {
	ID         int64     `json:"id"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   int64     `json:"to_user_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	// the user who sent it, only loaded when listing pending requests
	From *User `json:"-"`
}
