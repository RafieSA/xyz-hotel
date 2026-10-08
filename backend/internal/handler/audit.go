package handler

import (
	"strconv"

	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"

	"github.com/gofiber/fiber/v2"
)

// AuditHandler serves GET /api/admin/audit-logs
type AuditHandler struct {
	AuditRepo *repo.AuditRepo
}

func NewAuditHandler(ar *repo.AuditRepo) *AuditHandler {
	return &AuditHandler{AuditRepo: ar}
}

// ListAuditLogs handles GET /api/admin/audit-logs?limit&offset&entity&action
// RBAC owner/manager enforced via middleware; handler double-checks role.
func (h *AuditHandler) ListAuditLogs(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	if role != model.RoleOwner && role != model.RoleManager {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You do not have permission to view audit logs"})
	}
	if h.AuditRepo == nil || h.AuditRepo.DB == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "Service is temporarily unavailable. Please try again later"})
	}
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")
	entity := c.Query("entity")
	action := c.Query("action")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	list, err := h.AuditRepo.List(limit, offset, entity, action)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not load audit logs. Please try again"})
	}
	total, _ := h.AuditRepo.Count(entity, action)
	return c.JSON(fiber.Map{
		"data":   list,
		"limit":  limit,
		"offset": offset,
		"total":  total,
	})
}
