package domain

import "errors"

var (
	ErrInvalidUserID = errors.New("invalid user_id")
	ErrInvalidDate   = errors.New("invalid date")
	ErrEmptyTitle    = errors.New("event title is required")
	ErrEventNotFound = errors.New("event not found")
)
