package handler

import "github.com/gofiber/fiber/v2"

// ListBookingsStub placeholder for Fase 1.
func ListBookingsStub(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data":    []interface{}{},
		"message": "stub - not implemented yet",
	})
}

// CreateBookingStub placeholder for Fase 1.
func CreateBookingStub(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "stub - not implemented yet",
	})
}
