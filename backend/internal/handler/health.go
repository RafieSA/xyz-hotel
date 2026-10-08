package handler

import "github.com/gofiber/fiber/v2"

// Health returns service status.
func Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "xyz-hotel",
	})
}
