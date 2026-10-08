package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"

	"github.com/jmoiron/sqlx"
)

// AvailabilityResult holds availability check result.
type AvailabilityResult struct {
	RoomTypeID int64           `json:"room_type_id"`
	RoomType   *model.RoomType `json:"room_type,omitempty"`
	TotalUnits int             `json:"total_units"`
	Occupied   int             `json:"occupied"`
	Available  int             `json:"available"`
	CheckIn    string          `json:"check_in"`
	CheckOut   string          `json:"check_out"`
}

// AvailabilityService handles availability checks with ACID guarantees.
type AvailabilityService struct {
	DB          *sqlx.DB
	BookingRepo *repo.BookingRepo
	RoomRepo    *repo.RoomRepo
}

func NewAvailabilityService(db *sqlx.DB) *AvailabilityService {
	return &AvailabilityService{
		DB:          db,
		BookingRepo: repo.NewBookingRepo(db),
		RoomRepo:    repo.NewRoomRepo(db),
	}
}

// parseDate validates YYYY-MM-DD and ensures checkIn < checkOut.
func parseDate(checkIn, checkOut string) (time.Time, time.Time, error) {
	ci, err := time.Parse("2006-01-02", checkIn)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid check_in format, expected YYYY-MM-DD")
	}
	co, err := time.Parse("2006-01-02", checkOut)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid check_out format, expected YYYY-MM-DD")
	}
	if !co.After(ci) {
		return time.Time{}, time.Time{}, fmt.Errorf("check_out must be after check_in")
	}
	return ci, co, nil
}

// CheckAvailability returns remaining units for a room type and date range (public, no lock).
func (s *AvailabilityService) CheckAvailability(ctx context.Context, roomTypeID int64, checkIn, checkOut string) (*AvailabilityResult, error) {
	if _, _, err := parseDate(checkIn, checkOut); err != nil {
		return nil, err
	}
	rt, err := s.RoomRepo.GetRoomTypeByID(roomTypeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room type not found")
		}
		return nil, err
	}
	occupied, err := s.BookingRepo.CountOverlapping(roomTypeID, checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	available := rt.TotalUnits - occupied
	if available < 0 {
		available = 0
	}
	return &AvailabilityResult{
		RoomTypeID: roomTypeID,
		RoomType:   rt,
		TotalUnits: rt.TotalUnits,
		Occupied:   occupied,
		Available:  available,
		CheckIn:    checkIn,
		CheckOut:   checkOut,
	}, nil
}

// CreateBooking creates a booking transactionally with FOR UPDATE to prevent race conditions.
// Steps: BEGIN; SELECT room_type FOR UPDATE; SELECT COUNT(*) FROM bookings FOR UPDATE; validate availability; INSERT pending_payment with total_price snapshot; audit log; COMMIT.
func (s *AvailabilityService) CreateBooking(ctx context.Context, userID, roomTypeID int64, checkInStr, checkOutStr string, guests int) (*model.Booking, error) {
	if guests < 1 {
		return nil, fmt.Errorf("guests must be >=1")
	}
	ci, co, err := parseDate(checkInStr, checkOutStr)
	if err != nil {
		return nil, err
	}
	// Nights = ceil((checkOut - checkIn)/24h)
	nights := int(co.Sub(ci).Hours() / 24)
	if nights < 1 {
		nights = 1
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

	// Lock room_type row to serialize concurrent bookings for same type
	rt, err := s.RoomRepo.GetRoomTypeByIDTx(tx, roomTypeID)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room type not found")
		}
		return nil, err
	}
	if guests > rt.Capacity {
		_ = tx.Rollback()
		return nil, fmt.Errorf("guests exceeds room capacity (%d)", rt.Capacity)
	}

	occupied, err := s.BookingRepo.CountOverlappingTx(tx, roomTypeID, checkInStr, checkOutStr)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	available := rt.TotalUnits - occupied
	if available <= 0 {
		_ = tx.Rollback()
		return nil, fmt.Errorf("no available units for selected dates")
	}

	totalPrice := rt.Price * int64(nights)

	// Insert booking pending_payment
	b := &model.Booking{
		UserID:     userID,
		RoomTypeID: roomTypeID,
		CheckIn:    ci,
		CheckOut:   co,
		Guests:     guests,
		TotalPrice: totalPrice,
		Status:     model.BookingPendingPayment,
	}
	created, err := s.BookingRepo.CreateTx(tx, b)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Audit log
	payload, _ := json.Marshal(map[string]interface{}{
		"room_type_id": roomTypeID,
		"check_in":     checkInStr,
		"check_out":    checkOutStr,
		"guests":       guests,
		"total_price":  totalPrice,
	})
	_, err = tx.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, payload) VALUES ($1,$2,$3,$4,$5::jsonb)`,
		userID, "booking.create", "bookings", created.ID, string(payload))
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("audit log failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	slog.Info("booking created", "booking_id", created.ID, "user_id", userID, "room_type_id", roomTypeID, "nights", nights, "total_price", totalPrice)
	return created, nil
}
// IsOverlapping reports whether [aStart, aEnd) overlaps [bStart, bEnd).
// Overlap condition from SQL: aStart < bEnd && bStart < aEnd.
// Pure function for unit testing without DB.
func IsOverlapping(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

// CalculateAvailable computes remaining units given total and overlapping counts.
// Clamped to 0.
func CalculateAvailable(totalUnits, overlapping int) int {
	avail := totalUnits - overlapping
	if avail < 0 {
		return 0
	}
	return avail
}

// NightsBetween returns nights between checkIn and checkOut (YYYY-MM-DD).
func NightsBetween(checkIn, checkOut string) (int, error) {
	ci, co, err := parseDate(checkIn, checkOut)
	if err != nil {
		return 0, err
	}
	nights := int(co.Sub(ci).Hours() / 24)
	if nights < 1 {
		nights = 1
	}
	return nights, nil
}
