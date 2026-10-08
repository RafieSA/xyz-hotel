package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Claims holds JWT payload for legacy handler tokens (user_id).
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ServiceClaims mirrors service.Claims (uid) for compatibility with AuthService tokens.
type ServiceClaims struct {
	UserID    int64  `json:"uid"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// Auth validates JWT Bearer token and sets claims in context.
// Sets Locals: "user_id" (int64), "role" (string), "email" (string), "claims" (*Claims or *ServiceClaims)
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

		// Try service claims first (with uid + typ)
		sClaims := &ServiceClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, sClaims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})
		if err == nil && token.Valid {
			// Reject refresh tokens for auth-guarded routes (only access allowed)
			if sClaims.TokenType == "refresh" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token not allowed"})
			}
			if sClaims.TokenType != "" && sClaims.TokenType != "access" {
				// unknown typ, allow if not refresh but check expiry already done
			}
			c.Locals("user_id", sClaims.UserID)
			c.Locals("role", sClaims.Role)
			c.Locals("email", sClaims.Email)
			c.Locals("claims", sClaims)
			return c.Next()
		}

		// Fallback to legacy claims (user_id)
		claims := &Claims{}
		token2, err2 := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})
		if err2 != nil || !token2.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid or expired token"})
		}
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("email", claims.Email)
		c.Locals("claims", claims)
		return c.Next()
	}
}

// OptionalAuth tries to parse token if present, but doesn't block.
func OptionalAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return c.Next()
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Next()
		}
		sClaims := &ServiceClaims{}
		token, err := jwt.ParseWithClaims(parts[1], sClaims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err == nil && token.Valid && sClaims.TokenType != "refresh" {
			c.Locals("user_id", sClaims.UserID)
			c.Locals("role", sClaims.Role)
			c.Locals("email", sClaims.Email)
			c.Locals("claims", sClaims)
			return c.Next()
		}
		claims := &Claims{}
		token2, err2 := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err2 == nil && token2.Valid {
			c.Locals("user_id", claims.UserID)
			c.Locals("role", claims.Role)
			c.Locals("email", claims.Email)
			c.Locals("claims", claims)
		}
		return c.Next()
	}
}
