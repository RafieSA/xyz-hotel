package handler

import (
	"xyz-hotel/backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	AuthService *service.AuthService
	Validator   *validator.Validate
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: svc, Validator: validator.New()}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid body"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "validation failed", "details": err.Error()})
	}
	user, err := h.AuthService.Register(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if err.Error() == "email already registered" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": err.Error()})
		}
		if err.Error() == "password must be at least 8 characters" || err.Error() == "invalid email format" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	// issue tokens
	access, refresh, err := h.AuthService.IssueTokensForTest(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "token generation failed"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"user":          fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
			"access_token":  access,
			"refresh_token": refresh,
		},
	})
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid body"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "validation failed", "details": err.Error()})
	}
	access, refresh, user, err := h.AuthService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"user":          fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
			"access_token":  access,
			"refresh_token": refresh,
		},
	})
}

// Me GET /api/auth/me (auth required)
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	var uid int64
	switch v := uidVal.(type) {
	case int64:
		uid = v
	case int:
		uid = int64(v)
	case float64:
		uid = int64(v)
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	user, err := h.AuthService.UserRepo.FindByID(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "user not found"})
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role}})
}

// Refresh POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "refresh_token required"})
	}
	access, refresh, err := h.AuthService.Refresh(c.Context(), body.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"access_token": access, "refresh_token": refresh}})
}
