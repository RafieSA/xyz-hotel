package handler

import (
	"xyz-hotel/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ReviewHandler struct {
	Svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{Svc: svc}
}

type CreateReviewRequest struct {
	BookingID int64  `json:"booking_id" validate:"required"`
	Rating    int    `json:"rating" validate:"required,gte=1,lte=5"`
	Comment   string `json:"comment" validate:"max=500"`
}

func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	uid, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	var req CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid body"})
	}
	if req.BookingID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "booking_id required"})
	}
	if req.Rating < 1 || req.Rating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "rating must be between 1 and 5"})
	}
	if len(req.Comment) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "comment must be at most 500 characters"})
	}
	rv, err := h.Svc.CreateReview(c.Context(), uid, req.BookingID, req.Rating, req.Comment)
	if err != nil {
		switch err {
		case service.ErrRatingBounds:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		case service.ErrCommentTooLong:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		case service.ErrNotCheckedOut:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		case service.ErrForbidden:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": err.Error()})
		case service.ErrAlreadyReviewed:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": err.Error()})
		default:
			if err.Error() == "booking not found" {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to create review"})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": rv})
}

func (h *ReviewHandler) ListReviews(c *fiber.Ctx) error {
	roomTypeID, err := c.QueryParser(nil), error(nil)
	_ = roomTypeID
	_ = err
	id := c.QueryInt("room_type_id", 0)
	if id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "room_type_id query required"})
	}
	list, err := h.Svc.ListReviews(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed"})
	}
	return c.JSON(fiber.Map{"data": list})
}
