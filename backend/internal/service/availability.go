package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
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
	VoucherRepo *repo.VoucherRepo
}

func NewAvailabilityService(db *sqlx.DB) *AvailabilityService {
	return &AvailabilityService{
		DB:          db,
		BookingRepo: repo.NewBookingRepo(db),
		RoomRepo:    repo.NewRoomRepo(db),
		VoucherRepo: repo.NewVoucherRepo(db),
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
// Steps: BEGIN; SELECT room_type FOR UPDATE; SELECT COUNT(*) FROM bookings FOR UPDATE; validate availability; optional voucher validation with FOR UPDATE + increment; INSERT pending_payment with total_price snapshot (discounted if voucher); audit log; COMMIT.
func (s *AvailabilityService) CreateBooking(ctx context.Context, userID, roomTypeID int64, checkInStr, checkOutStr string, guests int, voucherCode string) (*model.Booking, error) {
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

	// Voucher handling (optional) — validate in same transaction with FOR UPDATE
	var voucherID *int64
	var discount float64
	totalPrice := rt.Price * int64(nights)
	trimmedCode := strings.TrimSpace(voucherCode)
	if trimmedCode != "" {
		// Ensure VoucherRepo is available (fallback to DB-backed repo if nil due to legacy init)
		vr := s.VoucherRepo
		if vr == nil {
			vr = repo.NewVoucherRepo(s.DB)
		}
		v, err := vr.FindByCodeForUpdate(tx, trimmedCode)
		if err != nil {
			_ = tx.Rollback()
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("voucher not found")
			}
			return nil, err
		}
		// Validate expiry, quota, min_nights
		now := time.Now()
		if v.ExpiresAt != nil && now.After(*v.ExpiresAt) {
			_ = tx.Rollback()
			return nil, fmt.Errorf("voucher expired")
		}
		if v.Quota != nil && v.UsedCount >= *v.Quota {
			_ = tx.Rollback()
			return nil, fmt.Errorf("voucher quota exceeded")
		}
		if nights < v.MinNights {
			_ = tx.Rollback()
			return nil, fmt.Errorf("voucher requires minimum %d nights", v.MinNights)
		}
		discount = v.Discount
		// Apply discount: price*nights*(100-discount)/100 with rounding
		discounted := int64(math.Round(float64(totalPrice) * (100 - discount) / 100))
		if discounted < 0 {
			discounted = 0
		}
		totalPrice = discounted
		// Increment used_count within same tx (FOR UPDATE lock ensures atomicity)
		if err := vr.IncrementUsedCountTx(tx, v.ID); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to increment voucher: %w", err)
		}
		idCopy := v.ID
		voucherID = &idCopy
		slog.Info("voucher applied", "code", trimmedCode, "voucher_id", v.ID, "discount", discount, "nights", nights, "original_total", rt.Price*int64(nights), "discounted_total", totalPrice)
	}

	// Insert booking pending_payment
	b := &model.Booking{
		UserID:     userID,
		RoomTypeID: roomTypeID,
		CheckIn:    ci,
		CheckOut:   co,
		Guests:     guests,
		TotalPrice: totalPrice,
		Status:     model.BookingPendingPayment,
		VoucherID:  voucherID,
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
		"voucher_code": trimmedCode,
		"voucher_id":   voucherID,
		"discount":     discount,
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
	slog.Info("booking created", "booking_id", created.ID, "user_id", userID, "room_type_id", roomTypeID, "nights", nights, "total_price", totalPrice, "voucher_code", trimmedCode)
	return created, nil
}

// CreateBookingLegacy is a compatibility shim for callers not passing voucherCode (deprecated, prefer CreateBooking with voucherCode).
func (s *AvailabilityService) CreateBookingLegacy(ctx context.Context, userID, roomTypeID int64, checkInStr, checkOutStr string, guests int) (*model.Booking, error) {
	return s.CreateBooking(ctx, userID, roomTypeID, checkInStr, checkOutStr, guests, "")
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
