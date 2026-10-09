package handler

import (
	"log/slog"
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
	if msg == "from must be before or equal to to" {
		msg = "Start date must be before or on end date"
	}
	if msg == "invalid from date" {
		msg = "Start date is invalid. Use YYYY-MM-DD format"
	}
	if msg == "invalid to date" {
		msg = "End date is invalid. Use YYYY-MM-DD format"
	}
	if containsReport(msg, "Start date is invalid") || containsReport(msg, "End date is invalid") || containsReport(msg, "Start date must be") || containsReport(msg, "invalid") || containsReport(msg, "from must be") || containsReport(msg, "Start date") || containsReport(msg, "End date") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": msg})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not generate the report. Please try again"})
}

func containsReport(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not calculate revenue. Please try again"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not calculate occupancy. Please try again"})
	}
	perDay, _ := h.Reports.ReportRepo.OccupancyPerDay(c.Context(), from, to)
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

// GetCalendar handles GET /api/admin/calendar?from&to (RBAC owner/manager/receptionist) returns 7 days x 18 units matrix
func (h *ReportHandler) GetCalendar(c *fiber.Ctx) error {
	fromStr, toStr := parseRange(c)
	result, err := h.Reports.GetCalendar(c.Context(), fromStr, toStr)
	if err != nil {
		slog.Error("calendar failed", "from", fromStr, "to", toStr, "err", err)
		return reportError(c, err)
	}
	return c.JSON(result)
}

// ExportCSV handles GET /api/admin/reports/export.csv?from&to (RBAC) returns CSV with header date,revenue,bookings,occupancy
func (h *ReportHandler) ExportCSV(c *fiber.Ctx) error {
	fromStr, toStr := parseRange(c)
	from, to, err := service.ParseReportRange(fromStr, toStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	data, err := h.Reports.ReportCSV(c.Context(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "We could not generate CSV. Please try again"})
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=report.csv")
	return c.Send(data)
}
