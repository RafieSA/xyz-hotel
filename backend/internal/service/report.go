package service

import (
	"context"
	"fmt"
	"time"

	"xyz-hotel/backend/internal/repo"

	"github.com/jmoiron/sqlx"
)

// ReportService orchestrates reporting calculations.
type ReportService struct {
	DB         *sqlx.DB
	ReportRepo *repo.ReportRepo
}

func NewReportService(db *sqlx.DB) *ReportService {
	return &ReportService{DB: db, ReportRepo: repo.NewReportRepo(db)}
}

// date range helper
func parseReportRange(fromStr, toStr string) (time.Time, time.Time, error) {
	now := time.Now()
	defaultTo := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	defaultFrom := defaultTo.AddDate(0, 0, -30)

	var from, to time.Time
	var err error
	if fromStr == "" {
		from = defaultFrom
	} else {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Start date is invalid. Use YYYY-MM-DD format")
		}
	}
	if toStr == "" {
		to = defaultTo
	} else {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("End date is invalid. Use YYYY-MM-DD format")
		}
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, fmt.Errorf("from must be before or equal to to")
	}
	return from, to, nil
}

// ParseReportRange exported for handler/tests.
func ParseReportRange(fromStr, toStr string) (time.Time, time.Time, error) {
	return parseReportRange(fromStr, toStr)
}

// GetOccupancyRate returns (occupied unit-days / total unit-days *100).
func (s *ReportService) GetOccupancyRate(ctx context.Context, from, to time.Time) (float64, error) {
	totalUnits, err := s.ReportRepo.TotalUnits(ctx)
	if err != nil {
		return 0, err
	}
	if totalUnits == 0 {
		return 0, nil
	}
	occupied, err := s.ReportRepo.OccupiedDays(ctx, from, to)
	if err != nil {
		return 0, err
	}
	// rangeDays inclusive count: to - from +1
	days := int(to.Sub(from).Hours()/24) + 1
	if days <= 0 {
		return 0, nil
	}
	totalUnitDays := int64(totalUnits * days)
	if totalUnitDays == 0 {
		return 0, nil
	}
	rate := float64(occupied) / float64(totalUnitDays) * 100
	return rate, nil
}

// GetRevenue delegates to repo.
func (s *ReportService) GetRevenue(ctx context.Context, from, to time.Time) (int64, error) {
	return s.ReportRepo.Revenue(ctx, from, to)
}

// Summary aggregates all report data.
type Summary struct {
	OccupancyRate   float64                `json:"occupancy_rate"`
	TotalRevenue    int64                  `json:"total_revenue"`
	TotalBookings   int64                  `json:"total_bookings"`
	ByStatus        map[string]int64       `json:"by_status"`
	ByRoomType      []repo.RoomTypeStat    `json:"by_room_type"`
	RevenuePerDay   []repo.DailyRevenue    `json:"revenue_per_day"`
	OccupancyPerDay []repo.DailyOccupancy `json:"occupancy_per_day"`
}

// GetBookingsStats returns aggregated stats for summary.
func (s *ReportService) GetBookingsStats(ctx context.Context, from, to time.Time) (*Summary, error) {
	totalUnits, err := s.ReportRepo.TotalUnits(ctx)
	if err != nil {
		return nil, err
	}
	occupiedDays, err := s.ReportRepo.OccupiedDays(ctx, from, to)
	if err != nil {
		return nil, err
	}
	days := int(to.Sub(from).Hours()/24) + 1
	var occRate float64
	if totalUnits > 0 && days > 0 {
		totalUnitDays := int64(totalUnits * days)
		occRate = float64(occupiedDays) / float64(totalUnitDays) * 100
	}

	revenue, err := s.ReportRepo.Revenue(ctx, from, to)
	if err != nil {
		return nil, err
	}
	total, err := s.ReportRepo.TotalBookings(ctx, from, to)
	if err != nil {
		return nil, err
	}
	byStatus, err := s.ReportRepo.CountByStatus(ctx, from, to)
	if err != nil {
		return nil, err
	}
	if byStatus == nil {
		byStatus = map[string]int64{}
	}
	// ensure all known statuses present as 0 for frontend convenience
	for _, k := range []string{"pending_payment", "waiting_verification", "verified", "checked_in", "checked_out", "rejected", "expired", "cancelled"} {
		if _, ok := byStatus[k]; !ok {
			byStatus[k] = 0
		}
	}
	byRoomType, err := s.ReportRepo.StatsByRoomType(ctx, from, to)
	if err != nil {
		return nil, err
	}
	revPerDay, err := s.ReportRepo.RevenuePerDay(ctx, from, to)
	if err != nil {
		return nil, err
	}
	occPerDay, err := s.ReportRepo.OccupancyPerDay(ctx, from, to)
	if err != nil {
		return nil, err
	}
	// enrich occupancy per day rate
	if totalUnits > 0 {
		for i := range occPerDay {
			occPerDay[i].Rate = float64(occPerDay[i].Occupied) / float64(totalUnits) * 100
		}
	}
	if revPerDay == nil {
		revPerDay = []repo.DailyRevenue{}
	}
	if occPerDay == nil {
		occPerDay = []repo.DailyOccupancy{}
	}
	if byRoomType == nil {
		byRoomType = []repo.RoomTypeStat{}
	}
	return &Summary{
		OccupancyRate:   occRate,
		TotalRevenue:    revenue,
		TotalBookings:   total,
		ByStatus:        byStatus,
		ByRoomType:      byRoomType,
		RevenuePerDay:   revPerDay,
		OccupancyPerDay: occPerDay,
	}, nil
}

// GetSummary convenience parsing strings.
func (s *ReportService) GetSummary(ctx context.Context, fromStr, toStr string) (*Summary, time.Time, time.Time, error) {
	from, to, err := parseReportRange(fromStr, toStr)
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}
	sum, err := s.GetBookingsStats(ctx, from, to)
	return sum, from, to, err
}
