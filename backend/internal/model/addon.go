package model

import "time"

// Addon represents addons catalog.
type Addon struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name" validate:"required"`
	Price       int64     `db:"price" json:"price" validate:"gte=0"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// BookingAddon represents booking_addons join with price snapshot.
type BookingAddon struct {
	BookingID int64     `db:"booking_id" json:"booking_id"`
	AddonID   int64     `db:"addon_id" json:"addon_id"`
	Price     int64     `db:"price" json:"price"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	// joined fields for convenience
	AddonName string `db:"addon_name" json:"addon_name,omitempty"`
}

// LoyaltyTx represents loyalty_tx ledger.
type LoyaltyTx struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	BookingID *int64    `db:"booking_id" json:"booking_id,omitempty"`
	Points    int       `db:"points" json:"points"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// RoomTypeImage represents room_type_images gallery.
type RoomTypeImage struct {
	ID         int64     `db:"id" json:"id"`
	RoomTypeID int64     `db:"room_type_id" json:"room_type_id"`
	URL        string    `db:"url" json:"url"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
