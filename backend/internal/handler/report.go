package handler

import (
	"strings"

	"xyz-hotel/backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// ReportHandler serves admin reporting endpoints.
type ReportHandler struct {
	Reports *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{Reports: svc}
}

func parseRange(c *fiber.Ctx) (string, string) {
	return strings.TrimSpace(c.Query("from")), strings.TrimSpace(c.Query("to"))
}

func reportError(c *fiber.Ctx, err error) error {
	msg := err.Error()
	if strings.Contains(msg, "invalid") || strings.Contains(msg, "from must be") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to generate report"})
}

// GetSummary handles GET /api/admin/reports/summary?from&to
func (h *ReportHandler) GetSummary(c *fiber.Ctx) error {
	fromStr, toStr := parseRange(c)
	sum, from, to, err := h.Reports.GetSummary(c.Context(), fromStr, toStr)
	if err != nil {
		return reportError(c, err)
	}
	return c.JSON(fiber.Map{
		"from":              from.Format("2006-01-02"),
		"to":                to.Format("2006-01-02"),
		"occupancy_rate":    sum.OccupancyRate,
		"total_revenue":     sum.TotalRevenue,
		"total_bookings":    sum.TotalBookings,
		"by_status":         sum.ByStatus,
		"by_room_type":      sum.ByRoomType,
		"revenue_per_day":   sum.RevenuePerDay,
		"occupancy_per_day": sum.OccupancyPerDay,
	})
}

// GetRevenue handles GET /api/admin/reports/revenue?from&to
func (h *ReportHandler) GetRevenue(c *fiber.Ctx) error {
	fromStr, toStr := parseRange(c)
	from, to, err := service.ParseReportRange(fromStr, toStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	rev, err := h.Reports.GetRevenue(c.Context(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to calculate revenue"})
	}
	perDay, _ := h.Reports.ReportRepo.RevenuePerDay(c.Context(), from, to)
	return c.JSON(fiber.Map{
		"from":            from.Format("2006-01-02"),
		"to":              to.Format("2006-01-02"),
		"total_revenue":   rev,
		"revenue_per_day": perDay,
	})
}

// GetOccupancy handles GET /api/admin/reports/occupancy?from&to
func (h *ReportHandler) GetOccupancy(c *fiber.Ctx) error {
	fromStr, toStr := parseRange(c)
	from, to, err := service.ParseReportRange(fromStr, toStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	rate, err := h.Reports.GetOccupancyRate(c.Context(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to calculate occupancy"})
	}
	perDay, _ := h.Reports.ReportRepo.OccupancyPerDay(c.Context(), from, to)
	// enrich rate per day
	totalUnits, _ := h.Reports.ReportRepo.TotalUnits(c.Context())
	for i := range perDay {
		if totalUnits > 0 {
			perDay[i].Rate = float64(perDay[i].Occupied) / float64(totalUnits) * 100
		}
	}
	return c.JSON(fiber.Map{
		"from":              from.Format("2006-01-02"),
		"to":                to.Format("2006-01-02"),
		"occupancy_rate":    rate,
		"occupancy_per_day": perDay,
	})
}
