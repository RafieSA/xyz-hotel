package model

import "time"

// Wishlist represents wishlists table.
type Wishlist struct {
	ID         int64     `db:"id" json:"id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	RoomTypeID int64     `db:"room_type_id" json:"room_type_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// WishlistWithRoomType joins wishlist with room type details.
type WishlistWithRoomType struct {
	ID         int64   `db:"id" json:"id"`
	UserID     int64   `db:"user_id" json:"user_id"`
	RoomTypeID int64   `db:"room_type_id" json:"room_type_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	RoomType   RoomTypeWithRating `json:"room_type"`
}
