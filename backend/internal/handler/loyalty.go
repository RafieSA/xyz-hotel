package handler

import (
	"xyz-hotel/backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// LoyaltyHandler handles loyalty points.
type LoyaltyHandler struct {
	Svc       *service.LoyaltyService
	Validator *validator.Validate
}

func NewLoyaltyHandler(svc *service.LoyaltyService) *LoyaltyHandler {
	return &LoyaltyHandler{Svc: svc, Validator: validator.New()}
}

// GetPoints handles GET /api/loyalty/points (Auth)
func (h *LoyaltyHandler) GetPoints(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in"})
	}
	pts, err := h.Svc.GetPoints(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load points"})
	}
	history, _ := h.Svc.GetHistory(userID)
	return c.JSON(fiber.Map{"data": fiber.Map{"loyalty_points": pts, "history": history}})
}

// RedeemRequest validates POST /api/loyalty/redeem {points}
type RedeemRequest struct {
	Points int `json:"points" validate:"required,gte=100"`
}

// Redeem handles POST /api/loyalty/redeem {points} -> discount; deducts points
func (h *LoyaltyHandler) Redeem(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in"})
	}
	var req RedeemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Points must be at least 100", "details": err.Error()})
	}
	discount, err := h.Svc.RedeemPoints(userID, req.Points)
	if err != nil {
		msg := err.Error()
		if contains(msg, "Requires 100 points") || contains(msg, "points must be") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
	}
	pts, _ := h.Svc.GetPoints(userID)
	return c.JSON(fiber.Map{"data": fiber.Map{"points_redeemed": req.Points, "discount": discount, "remaining_points": pts}})
}
