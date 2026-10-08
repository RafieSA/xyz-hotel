package model

import "time"

const (
	ReviewRatingMin = 1
	ReviewRatingMax = 5
	MaxCommentLen   = 500
)

// Review represents reviews table  -  one per booking after checked_out.
type Review struct {
	ID         int64     `db:"id" json:"id"`
	BookingID  int64     `db:"booking_id" json:"booking_id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	RoomTypeID int64     `db:"room_type_id" json:"room_type_id"`
	Rating     int       `db:"rating" json:"rating"`
	Comment    string    `db:"comment" json:"comment"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// RoomTypeWithRating extends RoomType with aggregated rating fields.
type RoomTypeWithRating struct {
	RoomType
	AvgRating   float64 `db:"avg_rating" json:"avg_rating"`
	ReviewCount int     `db:"review_count" json:"review_count"`
}
