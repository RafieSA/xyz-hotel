package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// InvoiceHandler serves invoice PDF generation.
type InvoiceHandler struct {
	InvoiceService *service.InvoiceService
	BookingGetter  interface {
		GetByID(id int64) (*model.Booking, error)
	}
}

func NewInvoiceHandler(svc *service.InvoiceService, getter interface {
	GetByID(id int64) (*model.Booking, error)
}) *InvoiceHandler {
	return &InvoiceHandler{InvoiceService: svc, BookingGetter: getter}
}

// GetInvoice handles GET /api/bookings/:id/invoice
// Auth required, owner of booking or admin, only if status verified/checked_in/checked_out else 409
func (h *InvoiceHandler) GetInvoice(c *fiber.Ctx) error {
	userID, ok := getAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to view invoices"})
	}
	role, _ := c.Locals("role").(string)
	idStr := c.Params("id")
	bookingID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invoice ID is invalid"})
	}
	if h.BookingGetter == nil || h.InvoiceService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
	}
	booking, err := h.BookingGetter.GetByID(bookingID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
	}
	isAdmin := role == model.RoleOwner || role == model.RoleManager || role == model.RoleReceptionist
	if !isAdmin && booking.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You do not have permission to view this invoice"})
	}
	allowed := map[string]bool{
		model.BookingVerified:   true,
		model.BookingCheckedIn:  true,
		model.BookingCheckedOut: true,
	}
	if !allowed[booking.Status] {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": fmt.Sprintf("Invoice is not available yet. Booking status is %s. Invoice requires verified, checked_in or checked_out", booking.Status)})
	}
	relPath, err := h.InvoiceService.GenerateInvoice(c.Context(), bookingID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not generate the invoice. Please try again"})
	}
	// Determine absolute path
	fullPath := relPath
	if !filepath.IsAbs(relPath) {
		fullPath = filepath.Join(".", relPath)
	}
	if _, err := os.Stat(fullPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Invoice file not found after generation"})
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="invoice-%d.pdf"`, bookingID))
	// Prefer attachment variant as well via header? keep inline for preview
	return c.SendFile(fullPath)
}
