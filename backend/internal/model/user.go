package model

import "time"

type User struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth string    `json:"date_of_birth"`
	Avatar      string    `json:"avatar"`
	Nickname    string    `json:"nickname"`
	AboutMe     string    `json:"about_me"`
	Private     bool      `json:"private"`
	CreatedAt   time.Time `json:"created_at"`
	// the two relations are relative to whoever the read was made by, and the
	// three counts are not: they come off the user_view, so they read the same
	// for everyone. The reads that take no viewer leave both relations at
	// FollowStateNone.
	IsFollowed  int64 `json:"is_followed"`  // FollowStateNone/Active/Pending: the viewer follows this user
	IsFollowing int64 `json:"is_following"` // FollowStateNone/Active/Pending: this user follows the viewer
	Followers   int64 `json:"followers"`    // accepted requests only, whoever the viewer is
	Following   int64 `json:"following"`    // accepted requests only, whoever the viewer is
	// PostCount is every non-group post the user has, of any privacy. A view
	// cannot be handed an id, so this cannot be narrowed to what one viewer may
	// read — the posts tab behind it may show fewer.
	PostCount int64 `json:"post_count"`
}
