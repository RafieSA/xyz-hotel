package repo

import "github.com/jmoiron/sqlx"

// BookingRepo stub.
type BookingRepo struct {
	DB *sqlx.DB
}

func NewBookingRepo(db *sqlx.DB) *BookingRepo { return &BookingRepo{DB: db} }

// CountOverlapping counts verified/checked_in bookings overlapping [checkIn, checkOut).
func (r *BookingRepo) CountOverlapping(roomTypeID int64, checkIn, checkOut string) (int, error) {
	// TODO: SELECT COUNT(*) FROM bookings WHERE room_type_id=$1 AND status IN ('verified','checked_in') AND check_in < $3 AND check_out > $2
	return 0, nil
}
