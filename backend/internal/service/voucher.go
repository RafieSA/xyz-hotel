package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

// VoucherService handles voucher validation and creation.
type VoucherService struct {
	Repo     *repo.VoucherRepo
	RoomRepo *repo.RoomRepo
}

func NewVoucherService(vr *repo.VoucherRepo, rr *repo.RoomRepo) *VoucherService {
	return &VoucherService{Repo: vr, RoomRepo: rr}
}

// CreateVoucher validates input and inserts a voucher (admin only).
func (s *VoucherService) CreateVoucher(ctx context.Context, v *model.Voucher) (*model.Voucher, error) {
	v.Code = strings.TrimSpace(v.Code)
	if v.Code == "" {
		return nil, fmt.Errorf("voucher code is required")
	}
	// Use validator for discount/min_nights constraints.
	validate := validator.New()
	if err := validate.Struct(v); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	if v.Discount < 0 || v.Discount > 100 {
		return nil, fmt.Errorf("discount must be between 0 and 100")
	}
	if v.Quota != nil && *v.Quota < 0 {
		return nil, fmt.Errorf("quota must be >=0")
	}

	created, err := s.Repo.Create(ctx, v)
	if err != nil {
		// Detect unique violation for code.
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "Duplicate") {
			return nil, fmt.Errorf("voucher code already exists")
		}
		return nil, err
	}
	slog.Info("voucher created", "id", created.ID, "code", created.Code, "discount", created.Discount)
	return created, nil
}

// ListVouchers returns all vouchers.
func (s *VoucherService) ListVouchers(ctx context.Context) ([]model.Voucher, error) {
	return s.Repo.List(ctx)
}

// ValidateResult holds voucher validation preview.
type ValidateResult struct {
	Voucher        *model.Voucher `json:"voucher"`
	Discount       float64        `json:"discount_percent"`
	VoucherID      int64          `json:"voucher_id"`
	Nights         int            `json:"nights"`
	OriginalTotal  int64          `json:"original_total"`
	DiscountedTotal int64         `json:"discounted_total"`
}

// ValidateAndApply validates voucher and computes discounted total (non-transactional preview).
// Checks: exists, not expired, quota not exceeded, min_nights.
func (s *VoucherService) ValidateAndApply(ctx context.Context, voucherCode string, roomTypeID int64, checkIn, checkOut string) (*ValidateResult, error) {
	code := strings.TrimSpace(voucherCode)
	if code == "" {
		return nil, fmt.Errorf("voucher code is required")
	}
	ci, co, err := parseDate(checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	nights := int(co.Sub(ci).Hours() / 24)
	if nights < 1 {
		nights = 1
	}

	v, err := s.Repo.FindByCode(ctx, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("voucher not found")
		}
		return nil, err
	}

	if err := validateVoucher(v, nights, time.Now()); err != nil {
		return nil, err
	}

	// Compute totals if room type provided.
	var originalTotal int64
	var discountedTotal int64
	if s.RoomRepo != nil && roomTypeID != 0 {
		rt, err := s.RoomRepo.GetRoomTypeByID(roomTypeID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("room type not found")
			}
			return nil, err
		}
		originalTotal = rt.Price * int64(nights)
		discountedTotal = CalculateDiscountedPrice(originalTotal, v.Discount)
	}

	return &ValidateResult{
		Voucher:         v,
		Discount:        v.Discount,
		VoucherID:       v.ID,
		Nights:          nights,
		OriginalTotal:   originalTotal,
		DiscountedTotal: discountedTotal,
	}, nil
}

// ValidateForBookingTx validates voucher inside a transaction with FOR UPDATE lock.
// Returns voucher and discount for booking creation. Caller must have started tx.
func (s *VoucherService) ValidateForBookingTx(tx *sqlx.Tx, voucherCode string, nights int) (*model.Voucher, error) {
	code := strings.TrimSpace(voucherCode)
	if code == "" {
		return nil, fmt.Errorf("voucher code is required")
	}
	v, err := s.Repo.FindByCodeForUpdate(tx, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("voucher not found")
		}
		return nil, err
	}
	if err := validateVoucher(v, nights, time.Now()); err != nil {
		return nil, err
	}
	return v, nil
}

// validateVoucher performs pure checks: expiry, quota, min_nights. Exported for testing.
func validateVoucher(v *model.Voucher, nights int, now time.Time) error {
	if v.ExpiresAt != nil && now.After(*v.ExpiresAt) {
		return fmt.Errorf("voucher expired")
	}
	if v.Quota != nil && v.UsedCount >= *v.Quota {
		return fmt.Errorf("voucher quota exceeded")
	}
	if nights < v.MinNights {
		return fmt.Errorf("voucher requires minimum %d nights", v.MinNights)
	}
	return nil
}

// CalculateDiscountedPrice applies discount percent to a total price.
// discount in 0..100. Clamped to >=0. Uses math.Round for fairness.
func CalculateDiscountedPrice(total int64, discount float64) int64 {
	if discount <= 0 {
		return total
	}
	if discount >= 100 {
		return 0
	}
	discounted := math.Round(float64(total) * (100 - discount) / 100)
	if discounted < 0 {
		return 0
	}
	return int64(discounted)
}

// DiscountedTotal computes discounted total from per-night price and nights.
func DiscountedTotal(pricePerNight int64, nights int, discount float64) int64 {
	base := pricePerNight * int64(nights)
	return CalculateDiscountedPrice(base, discount)
}
