package service

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// LoyaltyService handles earn/redeem logic: 10 points per night on checked_out only, 100 points = 100k discount.
type LoyaltyService struct {
	DB *sqlx.DB
}

func NewLoyaltyService(db *sqlx.DB) *LoyaltyService { return &LoyaltyService{DB: db} }

// GetPoints returns current loyalty_points for user.
func (s *LoyaltyService) GetPoints(userID int64) (int, error) {
	var pts int
	err := s.DB.Get(&pts, `SELECT loyalty_points FROM users WHERE id=$1`, userID)
	if err != nil {
		return 0, err
	}
	return pts, nil
}

// GetHistory returns loyalty_tx for user.
func (s *LoyaltyService) GetHistory(userID int64) ([]map[string]interface{}, error) {
	rows, err := s.DB.Queryx(`SELECT id, user_id, booking_id, points, created_at FROM loyalty_tx WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id, uid int64
		var bid *int64
		var pts int
		var ts time.Time
		if err := rows.Scan(&id, &uid, &bid, &pts, &ts); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{"id": id, "user_id": uid, "booking_id": bid, "points": pts, "created_at": ts})
	}
	if out == nil {
		out = []map[string]interface{}{}
	}
	return out, nil
}

// EarnPoints awards 10 points per night for a booking that is checked_out.
// Only checked_out bookings earn, idempotent via loyalty_tx uniqueness check.
func (s *LoyaltyService) EarnPoints(userID, bookingID int64) (int, error) {
	tx, err := s.DB.BeginTxx(nil, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// Fetch booking, verify owned and checked_out
	var b struct {
		ID       int64     `db:"id"`
		UserID   int64     `db:"user_id"`
		CheckIn  time.Time `db:"check_in"`
		CheckOut time.Time `db:"check_out"`
		Status   string    `db:"status"`
	}
	err = tx.Get(&b, `SELECT id, user_id, check_in, check_out, status FROM bookings WHERE id=$1 FOR UPDATE`, bookingID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if b.UserID != userID {
		_ = tx.Rollback()
		return 0, fmt.Errorf("not owner of booking")
	}
	if b.Status != "checked_out" {
		_ = tx.Rollback()
		return 0, fmt.Errorf("points are only earned after checked out")
	}
	// Check if already earned for this booking
	var existing int
	err = tx.Get(&existing, `SELECT COUNT(*) FROM loyalty_tx WHERE booking_id=$1 AND points > 0`, bookingID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if existing > 0 {
		_ = tx.Rollback()
		return 0, fmt.Errorf("points already earned for this booking")
	}
	nights := int(b.CheckOut.Sub(b.CheckIn).Hours() / 24)
	if nights < 1 {
		nights = 1
	}
	points := nights * 10

	if _, err := tx.Exec(`UPDATE users SET loyalty_points = loyalty_points + $1 WHERE id=$2`, points, userID); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if _, err := tx.Exec(`INSERT INTO loyalty_tx (user_id, booking_id, points) VALUES ($1,$2,$3)`, userID, bookingID, points); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return points, nil
}

// RedeemPoints deducts points and returns discount. 100 points = 100000 IDR.
func (s *LoyaltyService) RedeemPoints(userID int64, points int) (int64, error) {
	if points < 100 {
		return 0, fmt.Errorf("Requires 100 points, you have %d", points)
	}
	if points%100 != 0 {
		return 0, fmt.Errorf("points must be in multiples of 100")
	}
	tx, err := s.DB.BeginTxx(nil, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	var current int
	err = tx.Get(&current, `SELECT loyalty_points FROM users WHERE id=$1 FOR UPDATE`, userID)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, err
	}
	if current < points {
		_ = tx.Rollback()
		return 0, fmt.Errorf("Requires 100 points, you have %d", current)
	}
	discount := int64(points/100) * 100000
	if _, err := tx.Exec(`UPDATE users SET loyalty_points = loyalty_points - $1 WHERE id=$2`, points, userID); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if _, err := tx.Exec(`INSERT INTO loyalty_tx (user_id, points) VALUES ($1,$2)`, userID, -points); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return discount, nil
}

// ApplyLoyaltyDiscount calculates discount for given points without deducting; caller handles deduct in txn.
// Kept for availability integration: returns discount amount.
func (s *LoyaltyService) ApplyLoyaltyDiscount(points int) (int64, error) {
	if points < 100 {
		return 0, fmt.Errorf("Requires 100 points, you have %d", points)
	}
	if points%100 != 0 {
		return 0, fmt.Errorf("points must be in multiples of 100")
	}
	return int64(points/100) * 100000, nil
}

// DeductPointsTx deducts points inside an existing transaction, with audit row. Used by availability booking txn.
func (s *LoyaltyService) DeductPointsTx(tx *sqlx.Tx, userID int64, points int) (int64, error) {
	if points == 0 {
		return 0, nil
	}
	if points < 100 {
		return 0, fmt.Errorf("Requires 100 points, you have %d", points)
	}
	if points%100 != 0 {
		return 0, fmt.Errorf("points must be in multiples of 100")
	}
	var current int
	if err := tx.Get(&current, `SELECT loyalty_points FROM users WHERE id=$1 FOR UPDATE`, userID); err != nil {
		return 0, err
	}
	if current < points {
		return 0, fmt.Errorf("Requires 100 points, you have %d", current)
	}
	discount := int64(points/100) * 100000
	if _, err := tx.Exec(`UPDATE users SET loyalty_points = loyalty_points - $1 WHERE id=$2`, points, userID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`INSERT INTO loyalty_tx (user_id, points) VALUES ($1,$2)`, userID, -points); err != nil {
		return 0, err
	}
	return discount, nil
}
