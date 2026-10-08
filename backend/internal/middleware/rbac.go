package middleware

import "github.com/gofiber/fiber/v2"

// RequireRole checks that authenticated user has one of allowed roles.
// Fase 1 stub: reads role from Locals("role") set by Auth.
func RequireRole(allowed ...string) fiber.Handler {
	allow := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		allow[r] = true
	}
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role == "" {
			// In Fase 1 without DB, deny by default unless middleware is not mounted.
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden: missing role"})
		}
		if !allow[role] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden: insufficient role"})
		}
		return c.Next()
	}
}
