package handler

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// VoucherHandler handles voucher CRUD and validation.
type VoucherHandler struct {
	Vouchers  *service.VoucherService
	Validator *validator.Validate
}

func NewVoucherHandler(svc *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{Vouchers: svc, Validator: validator.New()}
}

// CreateVoucherRequest body for POST /api/admin/vouchers.
type CreateVoucherRequest struct {
	Code            string  `json:"code" validate:"required"`
	DiscountPercent float64 `json:"discount_percent" validate:"required,gte=0,lte=100"`
	MinNights       int     `json:"min_nights" validate:"gte=0"`
	Quota           *int    `json:"quota" validate:"omitempty,gte=0"`
	ExpiresAt       *string `json:"expires_at"` // ISO8601 optional, e.g. "2026-12-31T23:59:59Z"
}

// CreateVoucher handles POST /api/admin/vouchers (owner/manager).
func (h *VoucherHandler) CreateVoucher(c *fiber.Ctx) error {
	var req CreateVoucherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please check your voucher details and try again", "details": err.Error()})
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Voucher code is required"})
	}

	v := &model.Voucher{
		Code:      code,
		Discount:  req.DiscountPercent,
		MinNights: req.MinNights,
		Quota:     req.Quota,
	}
	if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
		// Try RFC3339 then date-only fallback.
		t, err := parseVoucherTime(strings.TrimSpace(*req.ExpiresAt))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Expiry date is invalid. Use YYYY-MM-DD or RFC3339 format"})
		}
		v.ExpiresAt = t
	}

	created, err := h.Vouchers.CreateVoucher(c.Context(), v)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already exists") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": msg})
		}
		if strings.Contains(msg, "required") || strings.Contains(msg, "Discount") || strings.Contains(msg, "Quota") || strings.Contains(msg, "Please check") || strings.Contains(msg, "discount") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		slog.Error("create voucher failed", "err", err, "code", code)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not create the voucher. Please try again"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": created})
}

// ListVouchers handles GET /api/admin/vouchers (owner/manager).
func (h *VoucherHandler) ListVouchers(c *fiber.Ctx) error {
	list, err := h.Vouchers.ListVouchers(c.Context())
	if err != nil {
		slog.Error("list vouchers failed", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load vouchers. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// ValidateVoucher handles GET /api/vouchers/validate?code&room_type_id&check_in&check_out (public).
func (h *VoucherHandler) ValidateVoucher(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Voucher code is required"})
	}
	roomTypeIDStr := c.Query("room_type_id")
	if roomTypeIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type is required to validate voucher"})
	}
	roomTypeID, err := strconv.ParseInt(roomTypeIDStr, 10, 64)
	if err != nil || roomTypeID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type is invalid"})
	}
	checkIn := strings.TrimSpace(c.Query("check_in"))
	checkOut := strings.TrimSpace(c.Query("check_out"))
	if checkIn == "" || checkOut == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Check in and check out dates are required. Use YYYY-MM-DD"})
	}

	result, err := h.Vouchers.ValidateAndApply(c.Context(), code, roomTypeID, checkIn, checkOut)
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "not found"):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": msg})
		case strings.Contains(msg, "expired"), strings.Contains(msg, "usage limit"), strings.Contains(msg, "quota exceeded"), strings.Contains(msg, "needs at least"), strings.Contains(msg, "minimum"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		case strings.Contains(msg, "Check in date is invalid"), strings.Contains(msg, "Check out date is invalid"), strings.Contains(msg, "Check out must be after"), strings.Contains(msg, "invalid check"), strings.Contains(msg, "must be after"):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		slog.Error("voucher validate failed", "err", err, "code", code)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"voucher_id":       result.VoucherID,
			"code":             result.Voucher.Code,
			"discount_percent": result.Discount,
			"nights":           result.Nights,
			"original_total":   result.OriginalTotal,
			"discounted_total": result.DiscountedTotal,
			"voucher":          result.Voucher,
		},
	})
}

func parseVoucherTime(s string) (*time.Time, error) {
	// Try RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t, nil
	}
	// Try date only
	if t, err := time.Parse("2006-01-02", s); err == nil {
		// End of day for date-only expiry
		end := t.Add(24*time.Hour - time.Nanosecond)
		return &end, nil
	}
	// Try datetime without TZ
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return &t, nil
	}
	return nil, fmt.Errorf("Expiry date is invalid. Use YYYY-MM-DD or RFC3339 format")
}
