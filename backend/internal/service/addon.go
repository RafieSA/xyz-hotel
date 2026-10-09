package service

import (
	"encoding/json"
	"fmt"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"

	"github.com/jmoiron/sqlx"
)

// AddonService handles addon catalog and attaching to bookings.
type AddonService struct {
	DB        *sqlx.DB
	AddonRepo *repo.AddonRepo
	BookRepo  *repo.BookingRepo
}

func NewAddonService(db *sqlx.DB, addonRepo *repo.AddonRepo, bookRepo *repo.BookingRepo) *AddonService {
	return &AddonService{DB: db, AddonRepo: addonRepo, BookRepo: bookRepo}
}

// List returns all addons.
func (s *AddonService) List() ([]model.Addon, error) {
	return s.AddonRepo.List()
}

// AddToBooking attaches addons to a booking, only if pending_payment or waiting_verification and owned.
func (s *AddonService) AddToBooking(bookingID, userID int64, addonIDs []int64) (*model.Booking, error) {
	if len(addonIDs) == 0 {
		return nil, fmt.Errorf("addon_ids required")
	}
	tx, err := s.DB.BeginTxx(nil, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	var b model.Booking
	err = tx.Get(&b, `SELECT id, user_id, room_type_id, check_in, check_out, guests, total_price, status, created_at, updated_at FROM bookings WHERE id=$1 FOR UPDATE`, bookingID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if b.UserID != userID {
		_ = tx.Rollback()
		return nil, fmt.Errorf("not owner of booking")
	}
	if b.Status != model.BookingPendingPayment && b.Status != model.BookingWaitingVerification {
		_ = tx.Rollback()
		return nil, fmt.Errorf("only pending or waiting bookings can add addons")
	}
	addons, err := s.AddonRepo.ListByIDs(addonIDs)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if len(addons) != len(addonIDs) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("one or more addons not found")
	}
	// Dedup check: fetch existing
	var existingCount int
	if err := tx.Get(&existingCount, `SELECT COUNT(*) FROM booking_addons WHERE booking_id=$1 AND addon_id = ANY($2)`, bookingID, addonIDs); err != nil {
		// fallback via manual check if pg not support ANY int array binding via sqlx
		_ = err
	}
	// Insert with ON CONFLICT DO NOTHING; calculate total increase only for new ones
	var newAddons []model.Addon
	for _, a := range addons {
		var cnt int
		_ = tx.Get(&cnt, `SELECT COUNT(*) FROM booking_addons WHERE booking_id=$1 AND addon_id=$2`, bookingID, a.ID)
		if cnt == 0 {
			newAddons = append(newAddons, a)
		}
	}
	var totalAdd int64
	for _, a := range newAddons {
		if _, err := tx.Exec(`INSERT INTO booking_addons (booking_id, addon_id, price) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, bookingID, a.ID, a.Price); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		totalAdd += a.Price
	}
	if totalAdd > 0 {
		if _, err := tx.Exec(`UPDATE bookings SET total_price = total_price + $1, updated_at = now() WHERE id=$2`, totalAdd, bookingID); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	// audit
	payload, _ := json.Marshal(map[string]interface{}{"addon_ids": addonIDs, "added": totalAdd})
	if _, err := tx.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`, userID, "booking.add_addons", "bookings", bookingID, string(payload)); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("audit failed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// fetch updated booking
	updated, err := s.BookRepo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
