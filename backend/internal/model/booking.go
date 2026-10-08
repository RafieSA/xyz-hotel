package model

import "time"

// Booking status constants.
const (
	BookingPendingPayment      = "pending_payment"
	BookingWaitingVerification = "waiting_verification"
	BookingVerified            = "verified"
	BookingCheckedIn           = "checked_in"
	BookingCheckedOut          = "checked_out"
	BookingRejected            = "rejected"
	BookingExpired             = "expired"
	BookingCancelled           = "cancelled"
)

// Booking represents bookings table.
type Booking struct {
	ID           int64      `db:"id" json:"id"`
	UserID       int64      `db:"user_id" json:"user_id" validate:"required"`
	RoomTypeID   int64      `db:"room_type_id" json:"room_type_id" validate:"required"`
	RoomUnitID   *int64     `db:"room_unit_id" json:"room_unit_id,omitempty"`
	CheckIn      time.Time  `db:"check_in" json:"check_in" validate:"required"`
	CheckOut     time.Time  `db:"check_out" json:"check_out" validate:"required,gtfield=CheckIn"`
	Guests       int        `db:"guests" json:"guests" validate:"required,gte=1"`
	TotalPrice   int64      `db:"total_price" json:"total_price" validate:"gte=0"` // snapshot at creation
	Status       string     `db:"status" json:"status" validate:"required,oneof=pending_payment waiting_verification verified checked_in checked_out rejected expired cancelled"`
	VoucherID    *int64     `db:"voucher_id" json:"voucher_id,omitempty"`
	ProofURL     *string    `db:"proof_url" json:"proof_url,omitempty"`
	RejectReason *string    `db:"reject_reason" json:"reject_reason,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
