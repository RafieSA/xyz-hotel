package handler

import (
	"xyz-hotel/backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AddonHandler handles addons catalog and booking addons.
type AddonHandler struct {
	Svc       *service.AddonService
	Validator *validator.Validate
}

func NewAddonHandler(svc *service.AddonService) *AddonHandler {
	return &AddonHandler{Svc: svc, Validator: validator.New()}
}

// ListAddons handles GET /api/addons public
func (h *AddonHandler) ListAddons(c *fiber.Ctx) error {
	list, err := h.Svc.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load addons"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// AddToBookingRequest validates POST /api/bookings/:id/addons
type AddToBookingRequest struct {
	AddonIDs []int64 `json:"addon_ids" validate:"required,min=1"`
}

// AddToBooking handles POST /api/bookings/:id/addons {addon_ids:[1,2]} (Auth owner of booking, only pending/waiting else 409)
func (h *AddonHandler) AddToBooking(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in"})
	}
	bookingID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid booking id"})
	}
	var req AddToBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please provide addon_ids", "details": err.Error()})
	}
	updated, err := h.Svc.AddToBooking(bookingID, userID, req.AddonIDs)
	if err != nil {
		msg := err.Error()
		if contains(msg, "not owner") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You do not own this booking"})
		}
		if contains(msg, "only pending or waiting") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": msg})
		}
		if contains(msg, "addon_ids required") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
	}
	return c.JSON(fiber.Map{"data": updated})
}

// ListBookingAddons handles GET /api/bookings/:id/addons (auth, owner or admin)
func (h *AddonHandler) ListBookingAddons(c *fiber.Ctx) error {
	bookingID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid booking id"})
	}
	list, err := h.Svc.AddonRepo.FindForBooking(bookingID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load booking addons"})
	}
	return c.JSON(fiber.Map{"data": list})
}
