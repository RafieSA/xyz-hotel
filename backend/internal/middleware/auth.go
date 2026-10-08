package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Auth stub: validates JWT Bearer token and sets user claims in context.
// Fase 1: only structure, real verification in next phase.
func Auth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "missing Authorization header"})
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid Authorization format"})
		}
		tokenStr := parts[1]
		_, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token"})
		}
		// TODO: set claims to c.Locals("user", claims)
		return c.Next()
	}
}
