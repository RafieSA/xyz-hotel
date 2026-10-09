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
		return time.Time{}, time.Time{}, fmt.Errorf("Check in date is invalid. Use YYYY-MM-DD format")
	}
	co, err := time.Parse("2006-01-02", checkOut)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Check out date is invalid. Use YYYY-MM-DD format")
	}
	if !co.After(ci) {
		return time.Time{}, time.Time{}, fmt.Errorf("Check out must be after check in. Please adjust your dates")
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
			return nil, fmt.Errorf("Room type not found. Please choose a valid room type")
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
		return nil, fmt.Errorf("Guests must be at least 1")
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
			return nil, fmt.Errorf("Room type not found. Please choose a valid room type")
		}
		return nil, err
	}
	if guests > rt.Capacity {
		_ = tx.Rollback()
		return nil, fmt.Errorf("Too many guests for this room. Maximum is %d", rt.Capacity)
	}

	occupied, err := s.BookingRepo.CountOverlappingTx(tx, roomTypeID, checkInStr, checkOutStr)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	available := rt.TotalUnits - occupied
	if available <= 0 {
		_ = tx.Rollback()
		return nil, fmt.Errorf("No rooms available for those dates. Try different dates or room type")
	}

	// Voucher handling (optional) - validate in same transaction with FOR UPDATE
	// Dynamic pricing: sum per night with weekend +20% for Fri/Sat
	var voucherID *int64
	var discount float64
	var totalPrice int64
	for d := ci; d.Before(co); d = d.AddDate(0, 0, 1) {
		nightly := rt.Price
		if isWeekend(d) {
			nightly = int64(math.Round(float64(rt.Price) * 1.2))
		}
		totalPrice += nightly
	}
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
				return nil, fmt.Errorf("Voucher code not found. Check the code and try again")
			}
			return nil, err
		}
		// Validate expiry, quota, min_nights
		now := time.Now()
		if v.ExpiresAt != nil && now.After(*v.ExpiresAt) {
			_ = tx.Rollback()
			return nil, fmt.Errorf("This voucher has expired. Try a different code")
		}
		if v.Quota != nil && v.UsedCount >= *v.Quota {
			_ = tx.Rollback()
			return nil, fmt.Errorf("This voucher quota exceeded. It has reached its usage limit. Try a different code")
		}
		if nights < v.MinNights {
			_ = tx.Rollback()
			return nil, fmt.Errorf("This voucher requires minimum %d nights to apply", v.MinNights)
		}
		discount = v.Discount
		// Apply discount: weekend total * (100-discount)/100 with rounding
		originalTotal := totalPrice
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
		slog.Info("voucher applied", "code", trimmedCode, "voucher_id", v.ID, "discount", discount, "nights", nights, "original_total", originalTotal, "discounted_total", totalPrice)
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
	// Email log only: non-blocking, warm template. Resolve recipient email via DB lookup with fallback.
	func() {
		to := fmt.Sprintf("user-%d@xyz-hotel.local", userID)
		if s.DB != nil {
			var email string
			if err := s.DB.Get(&email, `SELECT email FROM users WHERE id=$1`, userID); err == nil && strings.TrimSpace(email) != "" {
				to = email
			}
		}
		subject := BookingCreatedSubject(created.ID)
		body := BookingCreatedBody(created.ID, roomTypeID, checkInStr, checkOutStr, guests, totalPrice)
		SendAsync(to, subject, body)
	}()
	return created, nil
}
// CreateBookingWithLoyalty creates booking with optional voucher and loyalty points deduction.
// loyaltyPoints: 0 means none, otherwise must be >=100 and multiple of 100. Discount 100pts=100k.
func (s *AvailabilityService) CreateBookingWithLoyalty(ctx context.Context, userID, roomTypeID int64, checkInStr, checkOutStr string, guests int, voucherCode string, loyaltyPoints int) (*model.Booking, error) {
	if loyaltyPoints < 0 {
		return nil, fmt.Errorf("loyalty_points cannot be negative")
	}
	if loyaltyPoints == 0 {
		return s.CreateBooking(ctx, userID, roomTypeID, checkInStr, checkOutStr, guests, voucherCode)
	}
	if loyaltyPoints < 100 {
		return nil, fmt.Errorf("Requires 100 points, you have %d", loyaltyPoints)
	}
	if loyaltyPoints%100 != 0 {
		return nil, fmt.Errorf("points must be in multiples of 100")
	}
	if guests < 1 {
		return nil, fmt.Errorf("Guests must be at least 1")
	}
	ci, co, err := parseDate(checkInStr, checkOutStr)
	if err != nil {
		return nil, err
	}
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
	rt, err := s.RoomRepo.GetRoomTypeByIDTx(tx, roomTypeID)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Room type not found. Please choose a valid room type")
		}
		return nil, err
	}
	if guests > rt.Capacity {
		_ = tx.Rollback()
		return nil, fmt.Errorf("Too many guests for this room. Maximum is %d", rt.Capacity)
	}
	occupied, err := s.BookingRepo.CountOverlappingTx(tx, roomTypeID, checkInStr, checkOutStr)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	available := rt.TotalUnits - occupied
	if available <= 0 {
		_ = tx.Rollback()
		return nil, fmt.Errorf("No rooms available for those dates. Try different dates or room type")
	}
	var voucherID *int64
	var discount float64
	var totalPrice int64
	for d := ci; d.Before(co); d = d.AddDate(0, 0, 1) {
		nightly := rt.Price
		if isWeekend(d) {
			nightly = int64(math.Round(float64(rt.Price) * 1.2))
		}
		totalPrice += nightly
	}
	trimmedCode := strings.TrimSpace(voucherCode)
	if trimmedCode != "" {
		vr := s.VoucherRepo
		if vr == nil {
			vr = repo.NewVoucherRepo(s.DB)
		}
		v, err := vr.FindByCodeForUpdate(tx, trimmedCode)
		if err != nil {
			_ = tx.Rollback()
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("Voucher code not found. Check the code and try again")
			}
			return nil, err
		}
		now := time.Now()
		if v.ExpiresAt != nil && now.After(*v.ExpiresAt) {
			_ = tx.Rollback()
			return nil, fmt.Errorf("This voucher has expired. Try a different code")
		}
		if v.Quota != nil && v.UsedCount >= *v.Quota {
			_ = tx.Rollback()
			return nil, fmt.Errorf("This voucher quota exceeded. It has reached its usage limit. Try a different code")
		}
		if nights < v.MinNights {
			_ = tx.Rollback()
			return nil, fmt.Errorf("This voucher requires minimum %d nights to apply", v.MinNights)
		}
		discount = v.Discount
		discounted := int64(math.Round(float64(totalPrice) * (100 - discount) / 100))
		if discounted < 0 {
			discounted = 0
		}
		totalPrice = discounted
		if err := vr.IncrementUsedCountTx(tx, v.ID); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to increment voucher: %w", err)
		}
		idCopy := v.ID
		voucherID = &idCopy
	}
	ls := NewLoyaltyService(s.DB)
	loyaltyDiscount, err := ls.DeductPointsTx(tx, userID, loyaltyPoints)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	totalPrice -= loyaltyDiscount
	if totalPrice < 0 {
		totalPrice = 0
	}
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
	payload, _ := json.Marshal(map[string]interface{}{
		"room_type_id":     roomTypeID,
		"check_in":         checkInStr,
		"check_out":        checkOutStr,
		"guests":           guests,
		"total_price":      totalPrice,
		"voucher_code":     trimmedCode,
		"voucher_id":       voucherID,
		"discount":         discount,
		"loyalty_points":   loyaltyPoints,
		"loyalty_discount": loyaltyDiscount,
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
	slog.Info("booking created with loyalty", "booking_id", created.ID, "user_id", userID, "room_type_id", roomTypeID, "nights", nights, "total_price", totalPrice, "voucher_code", trimmedCode, "loyalty_points", loyaltyPoints, "loyalty_discount", loyaltyDiscount)
	func() {
		to := fmt.Sprintf("user-%d@xyz-hotel.local", userID)
		if s.DB != nil {
			var email string
			if err := s.DB.Get(&email, `SELECT email FROM users WHERE id=$1`, userID); err == nil && strings.TrimSpace(email) != "" {
				to = email
			}
		}
		subject := BookingCreatedSubject(created.ID)
		body := BookingCreatedBody(created.ID, roomTypeID, checkInStr, checkOutStr, guests, totalPrice)
		SendAsync(to, subject, body)
	}()
	return created, nil
}

// isWeekend checks if Fri (5) or Sat (6) — weekend pricing +20%
func isWeekend(t time.Time) bool {
	wd := t.Weekday()
	return wd == time.Friday || wd == time.Saturday
}

// CalculatePrice computes total price per night with weekend premium.
// For each night from checkIn inclusive to checkOut exclusive:
// if isWeekend(night) then price*1.2 else price. Sum all nights.
func CalculatePrice(roomTypePrice int64, checkIn, checkOut string) (int64, error) {
	ci, co, err := parseDate(checkIn, checkOut)
	if err != nil {
		return 0, err
	}
	var total int64
	for d := ci; d.Before(co); d = d.AddDate(0, 0, 1) {
		nightly := roomTypePrice
		if isWeekend(d) {
			nightly = int64(math.Round(float64(roomTypePrice) * 1.2))
		}
		total += nightly
	}
	return total, nil
}

// CalculatePriceWithDiscount computes weekend-aware price then applies discount percent (0..100).
func CalculatePriceWithDiscount(roomTypePrice int64, checkIn, checkOut string, discount float64) (int64, error) {
	total, err := CalculatePrice(roomTypePrice, checkIn, checkOut)
	if err != nil {
		return 0, err
	}
	if discount <= 0 {
		return total, nil
	}
	if discount > 100 {
		discount = 100
	}
	discounted := int64(math.Round(float64(total) * (100 - discount) / 100))
	if discounted < 0 {
		discounted = 0
	}
	return discounted, nil
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
