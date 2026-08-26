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
}
