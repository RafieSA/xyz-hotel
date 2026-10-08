package handler

import (
	"log/slog"
	"strconv"

	"xyz-hotel/backend/internal/repo"
	"xyz-hotel/backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// BookingHandler handles availability and bookings.
type BookingHandler struct {
	Availability *service.AvailabilityService
	BookingRepo  *repo.BookingRepo
	Validator    *validator.Validate
}

func NewBookingHandler(svc *service.AvailabilityService, br *repo.BookingRepo) *BookingHandler {
	return &BookingHandler{
		Availability: svc,
		BookingRepo:  br,
		Validator:    validator.New(),
	}
}

// ListBookingsStub placeholder for Fase 1 (kept for compatibility, not used).
func ListBookingsStub(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data":    []interface{}{},
		"message": "stub - not implemented yet",
	})
}

// CreateBookingStub placeholder for Fase 1 (kept for compatibility, not used).
func CreateBookingStub(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "stub - not implemented yet",
	})
}

// GetAvailability handles GET /api/availability?room_type_id&check_in&check_out (public).
func (h *BookingHandler) GetAvailability(c *fiber.Ctx) error {
	roomTypeIDStr := c.Query("room_type_id")
	checkIn := c.Query("check_in")
	checkOut := c.Query("check_out")

	if roomTypeIDStr == "" || checkIn == "" || checkOut == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "room_type_id, check_in (YYYY-MM-DD), check_out (YYYY-MM-DD) are required"})
	}
	roomTypeID, err := strconv.ParseInt(roomTypeIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid room_type_id"})
	}

	result, err := h.Availability.CheckAvailability(c.Context(), roomTypeID, checkIn, checkOut)
	if err != nil {
		// SQL no rows => 404
		if err.Error() == "room type not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
		}
		if err.Error() == "check_out must be after check_in" || err.Error() == "invalid check_in format, expected YYYY-MM-DD" || err.Error() == "invalid check_out format, expected YYYY-MM-DD" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		slog.Error("availability check failed", "err", err, "room_type_id", roomTypeID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal error"})
	}
	return c.JSON(fiber.Map{
		"data": result,
	})
}

// CreateBookingRequest validation struct.
type CreateBookingRequest struct {
	RoomTypeID int64  `json:"room_type_id" validate:"required"`
	CheckIn    string `json:"check_in" validate:"required"`
	CheckOut   string `json:"check_out" validate:"required"`
	Guests     int    `json:"guests" validate:"required,gte=1"`
}

// CreateBooking handles POST /api/bookings (auth required).
func (h *BookingHandler) CreateBooking(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	userID, ok := userIDVal.(int64)
	if !ok {
		// fallback for int / float64 from JWT
		switch v := userIDVal.(type) {
		case int:
			userID = int64(v)
			ok = true
		case int64:
			ok = true
		case float64:
			userID = int64(v)
			ok = true
		}
	}
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized: missing user"})
	}

	var req CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid JSON body"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "validation failed", "details": err.Error()})
	}

	booking, err := h.Availability.CreateBooking(c.Context(), userID, req.RoomTypeID, req.CheckIn, req.CheckOut, req.Guests)
	if err != nil {
		msg := err.Error()
		switch msg {
		case "room type not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": msg})
		case "no available units for selected dates":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": msg})
		}
		if msg == "check_out must be after check_in" || msg == "invalid check_in format, expected YYYY-MM-DD" || msg == "invalid check_out format, expected YYYY-MM-DD" || msg == "guests must be >=1" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		// capacity exceeded etc.
		if contains(msg, "guests exceeds") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		slog.Error("create booking failed", "err", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal error"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": booking})
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

// ListBookings handles GET /api/bookings (auth required, own or admin).
// Customers see only own bookings (IDOR prevention). Owner/manager/receptionist see all.
func (h *BookingHandler) ListBookings(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	roleVal, _ := c.Locals("role").(string)
	var userID int64
	switch v := userIDVal.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	isAdmin := roleVal == "owner" || roleVal == "manager" || roleVal == "receptionist"
	if isAdmin {
		// BFLA: admin can see all, but still logged
		list, err := h.BookingRepo.ListAll()
		if err != nil {
			slog.Error("list all bookings failed", "err", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal error"})
		}
		return c.JSON(fiber.Map{"data": list})
	}
	// Customer: only own bookings (BOLA/IDOR prevention: ignore any query param user_id)
	list, err := h.BookingRepo.ListByUser(userID)
	if err != nil {
		slog.Error("list user bookings failed", "err", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal error"})
	}
	return c.JSON(fiber.Map{"data": list})
}
