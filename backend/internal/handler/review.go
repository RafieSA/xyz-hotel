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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to leave a review"})
	}
	var req CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if req.BookingID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Booking ID is required"})
	}
	if req.Rating < 1 || req.Rating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Rating must be 1 to 5. Please choose a rating from 1 to 5"})
	}
	if len(req.Comment) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Comment is too long. Keep it to 500 characters or fewer"})
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
			if err.Error() == "Booking not found" || err.Error() == "booking not found" {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not create your review. Please try again"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Room type is required to list reviews"})
	}
	list, err := h.Svc.ListReviews(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load reviews. Please try again"})
	}
	return c.JSON(fiber.Map{"data": list})
}
