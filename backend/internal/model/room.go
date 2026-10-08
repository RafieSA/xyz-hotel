package model

import "time"

// RoomType represents room_types table.
type RoomType struct {
	ID          int64      `db:"id" json:"id"`
	Name        string     `db:"name" json:"name" validate:"required"`
	Description string     `db:"description" json:"description"`
	Capacity    int        `db:"capacity" json:"capacity" validate:"required,gte=1"`
	Price       int64      `db:"price" json:"price" validate:"required,gte=0"` // price per night in IDR (minor unit)
	TotalUnits  int        `db:"total_units" json:"total_units" validate:"gte=0"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// RoomUnit represents room_units (physical units).
type RoomUnit struct {
	ID         int64      `db:"id" json:"id"`
	RoomTypeID int64      `db:"room_type_id" json:"room_type_id" validate:"required"`
	Code       string     `db:"code" json:"code" validate:"required"` // e.g. DLX-201
	Status     string     `db:"status" json:"status" validate:"required,oneof=available occupied dirty maintenance"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// Room unit status constants.
const (
	RoomStatusAvailable   = "available"
	RoomStatusOccupied    = "occupied"
	RoomStatusDirty       = "dirty"
	RoomStatusMaintenance = "maintenance"
)
