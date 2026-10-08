package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"

	"github.com/jmoiron/sqlx"
)

// BookingOpsService handles check-in/out and room-unit status transitions with ACID guarantees.
type BookingOpsService struct {
	DB          *sqlx.DB
	BookingRepo *repo.BookingRepo
	RoomRepo    *repo.RoomRepo
}

func NewBookingOpsService(db *sqlx.DB, bookingRepo *repo.BookingRepo, roomRepo *repo.RoomRepo) *BookingOpsService {
	if bookingRepo == nil {
		bookingRepo = repo.NewBookingRepo(db)
	}
	if roomRepo == nil {
		roomRepo = repo.NewRoomRepo(db)
	}
	return &BookingOpsService{DB: db, BookingRepo: bookingRepo, RoomRepo: roomRepo}
}

// CheckIn validates booking status verified, picks first available room_unit for room_type, assigns it and marks occupied.
// Returns error if status != verified or no available units.
func (s *BookingOpsService) CheckIn(ctx context.Context, bookingID int64, actorID int64) (*model.Booking, error) {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	booking, err := s.BookingRepo.GetByIDTx(tx, bookingID)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Booking not found")
		}
		return nil, err
	}
	if booking.Status != model.BookingVerified {
		_ = tx.Rollback()
		return nil, fmt.Errorf("Only verified bookings can be checked in. Current status is %s", booking.Status)
	}
	if booking.RoomUnitID != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("This booking already has a room assigned")
	}

	unitID, err := s.RoomRepo.FindAvailableUnitForTypeTx(ctx, tx, booking.RoomTypeID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("No rooms available to assign right now. Mark a room as available and try again: %w", err)
	}

	// Update booking
	var updated model.Booking
	err = tx.GetContext(ctx, &updated,
		`UPDATE bookings SET room_unit_id=$1, status=$2, updated_at=now() WHERE id=$3
		 RETURNING id, user_id, room_type_id, room_unit_id, check_in, check_out, guests, total_price, status, voucher_id, proof_url, reject_reason, created_at, updated_at`,
		unitID, model.BookingCheckedIn, bookingID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("update booking failed: %w", err)
	}

	// Occupy unit
	if err := s.RoomRepo.UpdateStatusTx(ctx, tx, unitID, model.RoomStatusOccupied); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("update room unit failed: %w", err)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"booking_id":   bookingID,
		"room_unit_id": unitID,
		"prev_status":  model.BookingVerified,
		"new_status":   model.BookingCheckedIn,
	})
	_, _ = tx.ExecContext(ctx, `INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`,
		actorID, "booking.checkin", "bookings", bookingID, string(payload))

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	slog.Info("check-in success", "booking_id", bookingID, "unit_id", unitID, "actor", actorID)
	return &updated, nil
}

// CheckOut validates booking status checked_in, marks checked_out and room dirty.
func (s *BookingOpsService) CheckOut(ctx context.Context, bookingID int64, actorID int64) (*model.Booking, error) {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	booking, err := s.BookingRepo.GetByIDTx(tx, bookingID)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Booking not found")
		}
		return nil, err
	}
	if booking.Status != model.BookingCheckedIn {
		_ = tx.Rollback()
		return nil, fmt.Errorf("Only checked in bookings can be checked out. Current status is %s", booking.Status)
	}
	if booking.RoomUnitID == nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("This booking has no room assigned")
	}
	unitID := *booking.RoomUnitID

	var updated model.Booking
	err = tx.GetContext(ctx, &updated,
		`UPDATE bookings SET status=$1, updated_at=now() WHERE id=$2
		 RETURNING id, user_id, room_type_id, room_unit_id, check_in, check_out, guests, total_price, status, voucher_id, proof_url, reject_reason, created_at, updated_at`,
		model.BookingCheckedOut, bookingID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("update booking failed: %w", err)
	}

	if err := s.RoomRepo.UpdateStatusTx(ctx, tx, unitID, model.RoomStatusDirty); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("update room unit to dirty failed: %w", err)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"booking_id":   bookingID,
		"room_unit_id": unitID,
		"prev_status":  model.BookingCheckedIn,
		"new_status":   model.BookingCheckedOut,
	})
	_, _ = tx.ExecContext(ctx, `INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`,
		actorID, "booking.checkout", "bookings", bookingID, string(payload))

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	slog.Info("check-out success", "booking_id", bookingID, "unit_id", unitID, "actor", actorID)
	return &updated, nil
}

// IsValidRoomStatusTransition checks allowed transitions for housekeeping.
// Allowed: available->occupied|dirty|maintenance, occupied->dirty, dirty->available|maintenance, maintenance->available
func IsValidRoomStatusTransition(from, to string) bool {
	if from == to {
		return false
	}
	allowed := map[string]map[string]bool{
		model.RoomStatusAvailable:   {model.RoomStatusOccupied: true, model.RoomStatusDirty: true, model.RoomStatusMaintenance: true},
		model.RoomStatusOccupied:    {model.RoomStatusDirty: true},
		model.RoomStatusDirty:       {model.RoomStatusAvailable: true, model.RoomStatusMaintenance: true},
		model.RoomStatusMaintenance: {model.RoomStatusAvailable: true},
	}
	if m, ok := allowed[from]; ok {
		return m[to]
	}
	return false
}

// ValidRoomStatuses set for validator.
var ValidRoomStatuses = map[string]bool{
	model.RoomStatusAvailable: true, model.RoomStatusOccupied: true, model.RoomStatusDirty: true, model.RoomStatusMaintenance: true,
}

// UpdateRoomUnitStatus updates room unit status with validation and audit. Uses transaction with FOR UPDATE.
func (s *BookingOpsService) UpdateRoomUnitStatus(ctx context.Context, unitID int64, newStatus string, actorID int64) (*model.RoomUnit, error) {
	if !ValidRoomStatuses[newStatus] {
		return nil, fmt.Errorf("Status is invalid. Choose available, occupied, dirty or maintenance")
	}
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	unit, err := s.RoomRepo.GetUnitByIDTx(ctx, tx, unitID)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Room not found")
		}
		return nil, err
	}
	if unit.Status == newStatus {
		_ = tx.Rollback()
		return nil, fmt.Errorf("Room is already %s", newStatus)
	}
	if !IsValidRoomStatusTransition(unit.Status, newStatus) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("Cannot change room from %s to %s", unit.Status, newStatus)
	}
	if err := s.RoomRepo.UpdateStatusTx(ctx, tx, unitID, newStatus); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"unit_id":     unitID,
		"prev_status": unit.Status,
		"new_status":  newStatus,
	})
	_, _ = tx.ExecContext(ctx, `INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`,
		actorID, "room_unit.status_update", "room_units", unitID, string(payload))

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	slog.Info("room unit status updated", "unit_id", unitID, "from", unit.Status, "to", newStatus, "actor", actorID)
	unit.Status = newStatus
	return unit, nil
}

// CancelBooking cancels a booking if caller owns it and status is pending_payment or waiting_verification.
// Sets status to cancelled, writes audit_logs and logs email asynchronously.
func (s *BookingOpsService) CancelBooking(ctx context.Context, bookingID, actorID int64) (*model.Booking, error) {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	booking, err := s.BookingRepo.GetByIDTx(tx, bookingID)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Booking not found")
		}
		return nil, err
	}
	if booking.UserID != actorID {
		_ = tx.Rollback()
		return nil, fmt.Errorf("forbidden: not owner")
	}
	if booking.Status != model.BookingPendingPayment && booking.Status != model.BookingWaitingVerification {
		_ = tx.Rollback()
		return nil, fmt.Errorf("conflict: only pending_payment or waiting_verification can be cancelled, current is %s", booking.Status)
	}

	var updated model.Booking
	err = tx.GetContext(ctx, &updated,
		`UPDATE bookings SET status=$1, updated_at=now() WHERE id=$2
		 RETURNING id, user_id, room_type_id, room_unit_id, check_in, check_out, guests, total_price, status, voucher_id, proof_url, reject_reason, created_at, updated_at`,
		model.BookingCancelled, bookingID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("cancel update failed: %w", err)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"booking_id":  bookingID,
		"prev_status": booking.Status,
		"new_status":  model.BookingCancelled,
		"voucher_id":  booking.VoucherID,
	})
	_, _ = tx.ExecContext(ctx, `INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`,
		actorID, "booking.cancel", "bookings", bookingID, string(payload))

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	slog.Info("booking cancelled", "booking_id", bookingID, "actor", actorID, "prev_status", booking.Status)
	// Email log only non-blocking; resolve recipient
	func() {
		to := fmt.Sprintf("user-%d@xyz-hotel.local", actorID)
		if s.DB != nil {
			var email string
			if err := s.DB.Get(&email, `SELECT email FROM users WHERE id=$1`, actorID); err == nil && email != "" {
				to = email
			}
		}
		subject := BookingCancelledSubject(bookingID)
		body := BookingCancelledBody(bookingID)
		SendAsync(to, subject, body)
	}()
	// voucher decrement note: logged via audit payload, no used_count rollback (business keeps quota consumed)
	slog.Info("booking cancel voucher logged", "booking_id", bookingID, "voucher_id", booking.VoucherID)
	return &updated, nil
}
