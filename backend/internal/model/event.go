package model

import "time"

const (
	EventChoiceGoing    = "going"
	EventChoiceNotGoing = "not_going"

	NotificationEventCreated = "event_created"
)

type GroupEvent struct {
	ID          int64     `json:"id"`
	GroupID     int64     `json:"group_id"`
	CreatorID   int64     `json:"creator_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DateTime    time.Time `json:"date_time"`
	CreatedAt   time.Time `json:"created_at"`
}

type EventResponse struct {
	ID        int64     `json:"id"`
	EventID   int64     `json:"event_id"`
	UserID    int64     `json:"user_id"`
	Choice    string    `json:"choice"`
	CreatedAt time.Time `json:"created_at"`
}

// EventListItem is one row of GET /groups/{id}/events: the event plus the
// response counts and the viewer's own choice ("", "going" or "not_going").
type EventListItem struct {
	GroupEvent
	CreatorFirstName string `json:"creator_first_name"`
	CreatorLastName  string `json:"creator_last_name"`
	GoingCount       int    `json:"going_count"`
	NotGoingCount    int    `json:"not_going_count"`
	MyChoice         string `json:"my_choice"`
}
