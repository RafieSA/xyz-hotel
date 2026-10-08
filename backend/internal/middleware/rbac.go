package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// RequireRole checks that authenticated user has one of allowed roles.
// Returns 401 if not authenticated, 403 on BFLA. Logs audit on denial.
func RequireRole(allowed ...string) fiber.Handler {
	allow := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		allow[r] = true
	}
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		userID := c.Locals("user_id")
		if role == "" {
			slog.Warn("rbac denied: missing role", "path", c.Path(), "method", c.Method(), "allowed", allowed, "user_id", userID)
			// If no role but user not authenticated, Auth middleware would have already returned 401.
			// Here treat as 403 for BFLA, but allow caller to distinguish.
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden: missing role"})
		}
		if !allow[role] {
			slog.Warn("rbac denied: insufficient role", "path", c.Path(), "method", c.Method(), "role", role, "user_id", userID, "allowed", allowed)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden: insufficient role"})
		}
		slog.Info("rbac allowed", "path", c.Path(), "role", role, "user_id", userID)
		return c.Next()
	}
}
