package model

import "time"

// Message represents messages table for in-app chat.
type Message struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	BookingID *int64    `db:"booking_id" json:"booking_id,omitempty"`
	Message   string    `db:"message" json:"message"`
	IsAdmin   bool      `db:"is_admin" json:"is_admin"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Conversation groups messages by user for admin inbox.
type Conversation struct {
	UserID      int64     `db:"user_id" json:"user_id"`
	UserName    string    `db:"user_name" json:"user_name"`
	UserEmail   string    `db:"user_email" json:"user_email"`
	LastMessage string    `db:"last_message" json:"last_message"`
	LastAt      time.Time `db:"last_at" json:"last_at"`
	Count       int       `db:"count" json:"count"`
}
