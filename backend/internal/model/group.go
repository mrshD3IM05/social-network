package model

import "time"

type Group struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	OwnerID     int64     `json:"owner_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
