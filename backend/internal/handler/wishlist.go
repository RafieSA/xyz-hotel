package handler

import (
	"strconv"

	"xyz-hotel/backend/internal/repo"
	"xyz-hotel/backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// WishlistHandler handles POST /api/wishlist/toggle, GET /api/wishlist, DELETE /api/wishlist/:room_type_id
type WishlistHandler struct {
	Service   *service.WishlistService
	Validator *validator.Validate
}

func NewWishlistHandler(svc *service.WishlistService) *WishlistHandler {
	return &WishlistHandler{Service: svc, Validator: validator.New()}
}

type wishlistToggleReq struct {
	RoomTypeID int64 `json:"room_type_id" validate:"required,gt=0"`
}

// Toggle handles POST /api/wishlist/toggle {room_type_id} (Auth)
func (h *WishlistHandler) Toggle(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to manage wishlist"})
	}
	var req wishlistToggleReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please provide a valid room type", "details": err.Error()})
	}
	added, err := h.Service.Toggle(userID, req.RoomTypeID)
	if err != nil {
		if repo.IsRoomTypeNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Room type not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not update your wishlist. Please try again"})
	}
	if added {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Added to wishlist", "added": true, "room_type_id": req.RoomTypeID})
	}
	return c.JSON(fiber.Map{"message": "Removed from wishlist", "added": false, "room_type_id": req.RoomTypeID})
}

// List handles GET /api/wishlist (Auth) -> enriched list with room_type for heart display
func (h *WishlistHandler) List(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to view wishlist"})
	}
	list, err := h.Service.ListEnriched(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load your wishlist. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}

// Delete handles DELETE /api/wishlist/:room_type_id (Auth) -> IDOR via user_id
func (h *WishlistHandler) Delete(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to manage wishlist"})
	}
	idStr := c.Params("room_type_id")
	if idStr == "" {
		idStr = c.Params("id")
	}
	roomTypeID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || roomTypeID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type ID is invalid"})
	}
	deleted, err := h.Service.Delete(userID, roomTypeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not remove from wishlist. Please try again"})
	}
	if !deleted {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Wishlist entry not found"})
	}
	return c.JSON(fiber.Map{"message": "Removed from wishlist", "room_type_id": roomTypeID})
}
