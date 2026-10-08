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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please check your name, email and password and try again", "details": err.Error()})
	}
	user, err := h.AuthService.Register(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		msg := err.Error()
		if msg == "This email is already registered. Please sign in instead" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": msg})
		}
		if msg == "Password needs at least 8 characters" || msg == "Enter a valid email address like name@example.com" || msg == "Name, email and password are required. Please fill in all fields" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
		}
		// fallback for legacy exact strings if service not yet updated
		if msg == "email already registered" || msg == "password must be at least 8 characters" || msg == "invalid email format" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not create your account. Please try again"})
	}
	// issue tokens
	access, refresh, err := h.AuthService.IssueTokensForTest(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not create your session. Please try again"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "We could not read your request. Check the format and try again"})
	}
	if err := h.Validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Please enter a valid email and password", "details": err.Error()})
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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Please sign in to continue"})
	}
	user, err := h.AuthService.UserRepo.FindByID(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "We could not find your account. Please sign in again"})
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role}})
}

// Refresh POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Refresh token is required. Please sign in again"})
	}
	access, refresh, err := h.AuthService.Refresh(c.Context(), body.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"access_token": access, "refresh_token": refresh}})
}
