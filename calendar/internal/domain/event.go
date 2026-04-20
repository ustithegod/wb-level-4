package domain

import "time"

type Event struct {
	ID       int64      `json:"id"`
	UserID   int64      `json:"user_id"`
	Date     time.Time  `json:"date"`
	Title    string     `json:"title"`
	RemindAt *time.Time `json:"remind_at,omitempty"`
	Created  time.Time  `json:"created_at"`
	Updated  time.Time  `json:"updated_at"`
}
