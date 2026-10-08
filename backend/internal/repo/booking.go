package repo

import (
	"xyz-hotel/backend/internal/model"

	"github.com/jmoiron/sqlx"
)

// BookingRepo handles bookings queries with $1 placeholders.
type BookingRepo struct {
	DB *sqlx.DB
}

func NewBookingRepo(db *sqlx.DB) *BookingRepo { return &BookingRepo{DB: db} }

// CountOverlapping counts verified/checked_in bookings overlapping [checkIn, checkOut).
// Overlap condition: check_in < $3 AND check_out > $2
func (r *BookingRepo) CountOverlapping(roomTypeID int64, checkIn, checkOut string) (int, error) {
	var count int
	err := r.DB.Get(&count,
		`SELECT COUNT(*) FROM bookings
		 WHERE room_type_id=$1
		   AND status IN ('verified','checked_in')
		   AND check_in < $3::date AND check_out > $2::date`,
		roomTypeID, checkIn, checkOut)
	return count, err
}

// CountOverlappingTx counts overlapping bookings inside a transaction with FOR UPDATE row locking.
// The SELECT ... FOR UPDATE ensures race-condition safety during booking creation.
// Postgres does not allow FOR UPDATE with aggregate, so we SELECT ids FOR UPDATE and count in Go.
func (r *BookingRepo) CountOverlappingTx(tx *sqlx.Tx, roomTypeID int64, checkIn, checkOut string) (int, error) {
	var ids []int64
	err := tx.Select(&ids,
		`SELECT id FROM bookings
		 WHERE room_type_id=$1
		   AND status IN ('verified','checked_in')
		   AND check_in < $3::date AND check_out > $2::date
		 FOR UPDATE`,
		roomTypeID, checkIn, checkOut)
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

// CountOverlappingAll counts overlapping bookings for any blocking status (used for display if needed).
func (r *BookingRepo) CountOverlappingAll(roomTypeID int64, checkIn, checkOut string) (int, error) {
	var count int
	err := r.DB.Get(&count,
		`SELECT COUNT(*) FROM bookings
		 WHERE room_type_id=$1
		   AND status IN ('pending_payment','waiting_verification','verified','checked_in')
		   AND check_in < $3::date AND check_out > $2::date`,
		roomTypeID, checkIn, checkOut)
	return count, err
}

// Create inserts a booking inside a transaction and returns the created booking.
func (r *BookingRepo) CreateTx(tx *sqlx.Tx, b *model.Booking) (*model.Booking, error) {
	var created model.Booking
	err := tx.Get(&created,
		`INSERT INTO bookings (user_id, room_type_id, check_in, check_out, guests, total_price, status)
		 VALUES ($1,$2,$3::date,$4::date,$5,$6,$7)
		 RETURNING id, user_id, room_type_id, room_unit_id, check_in, check_out, guests, total_price, status, voucher_id, proof_url, reject_reason, created_at, updated_at`,
		b.UserID, b.RoomTypeID, b.CheckIn.Format("2006-01-02"), b.CheckOut.Format("2006-01-02"), b.Guests, b.TotalPrice, b.Status)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// ListByUser returns bookings for a specific user (IDOR-safe).
func (r *BookingRepo) ListByUser(userID int64) ([]model.Booking, error) {
	var out []model.Booking
	err := r.DB.Select(&out,
		`SELECT id, user_id, room_type_id, room_unit_id, check_in, check_out, guests, total_price, status, voucher_id, proof_url, reject_reason, created_at, updated_at
		 FROM bookings WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Booking{}
	}
	return out, nil
}

// ListAll returns all bookings (for admin roles).
func (r *BookingRepo) ListAll() ([]model.Booking, error) {
	var out []model.Booking
	err := r.DB.Select(&out,
		`SELECT id, user_id, room_type_id, room_unit_id, check_in, check_out, guests, total_price, status, voucher_id, proof_url, reject_reason, created_at, updated_at
		 FROM bookings ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Booking{}
	}
	return out, nil
}

// ExpirePending sets bookings to expired where pending_payment older than 12 hours.
func (r *BookingRepo) ExpirePending() (int64, error) {
	res, err := r.DB.Exec(
		`UPDATE bookings SET status='expired', updated_at=now()
		 WHERE status='pending_payment' AND created_at < now() - interval '12 hours'`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// GetByID fetches a booking by id.
func (r *BookingRepo) GetByID(id int64) (*model.Booking, error) {
	var b model.Booking
	err := r.DB.Get(&b,
		`SELECT id, user_id, room_type_id, room_unit_id, check_in, check_out, guests, total_price, status, voucher_id, proof_url, reject_reason, created_at, updated_at
		 FROM bookings WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
