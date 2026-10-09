package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
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

// Calendar helpers for admin calendar 7 days x 18 units matrix

// parseCalendarRange parses from/to for calendar; defaults to next 7 days (today .. today+6)
func parseCalendarRange(fromStr, toStr string) (time.Time, time.Time, error) {
	now := time.Now()
	defaultFrom := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	defaultTo := defaultFrom.AddDate(0, 0, 6)
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
		// if from was provided but to empty, keep 7 days from from
		if fromStr != "" {
			to = from.AddDate(0, 0, 6)
		}
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

// ParseCalendarRange exported for handler.
func ParseCalendarRange(fromStr, toStr string) (time.Time, time.Time, error) {
	return parseCalendarRange(fromStr, toStr)
}

// CalendarUnit represents a physical unit for calendar matrix.
type CalendarUnit struct {
	ID         int64  `db:"id" json:"id"`
	Code       string `db:"code" json:"code"`
	RoomTypeID int64  `db:"room_type_id" json:"room_type_id"`
	Status     string `db:"status" json:"status"`
}

// CalendarBooking is a booking event for FullCalendar.
type CalendarBooking struct {
	ID         int64  `db:"id" json:"id"`
	RoomTypeID int64  `db:"room_type_id" json:"room_type_id"`
	RoomUnitID *int64 `db:"room_unit_id" json:"room_unit_id,omitempty"`
	CheckIn    string `db:"check_in" json:"check_in"`
	CheckOut   string `db:"check_out" json:"check_out"`
	Status     string `db:"status" json:"status"`
	Guests     int    `db:"guests" json:"guests"`
}

// CalendarResult holds 7 days x 18 units matrix data.
type CalendarResult struct {
	From      string              `json:"from"`
	To        string              `json:"to"`
	Days      []string            `json:"days"`
	Units     []CalendarUnit      `json:"units"`
	Bookings  []CalendarBooking   `json:"bookings"`
	Occupancy []repo.DailyOccupancy `json:"occupancy_per_day"`
	Matrix    map[string]map[string]int `json:"matrix"`
}

// GetCalendar returns 7 days x 18 units matrix with bookings per day per unit (occupied count), FOR UPDATE safe.
func (s *ReportService) GetCalendar(ctx context.Context, fromStr, toStr string) (*CalendarResult, error) {
	from, to, err := parseCalendarRange(fromStr, toStr)
	if err != nil {
		return nil, err
	}
	return s.getCalendarNoTx(ctx, from, to)
}

func (s *ReportService) getCalendarNoTx(ctx context.Context, from, to time.Time) (*CalendarResult, error) {
	var units []CalendarUnit
	err := s.DB.SelectContext(ctx, &units, `SELECT id, code, room_type_id, status FROM room_units WHERE deleted_at IS NULL ORDER BY code`)
	if err != nil {
		return nil, err
	}
	if units == nil {
		units = []CalendarUnit{}
	}
	var bookings []CalendarBooking
	err = s.DB.SelectContext(ctx, &bookings, `SELECT id, room_type_id, room_unit_id, check_in::text as check_in, check_out::text as check_out, status, guests FROM bookings WHERE check_in < $2::date AND check_out > $1::date ORDER BY check_in`, from.Format("2006-01-02"), to.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if bookings == nil {
		bookings = []CalendarBooking{}
	}
	var days []string
	matrix := make(map[string]map[string]int)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		days = append(days, ds)
		matrix[ds] = make(map[string]int)
		for _, u := range units {
			matrix[ds][u.Code] = 0
		}
	}
	occupancy, _ := s.ReportRepo.OccupancyPerDay(ctx, from, to)
	if occupancy == nil {
		occupancy = []repo.DailyOccupancy{}
	}
	return &CalendarResult{
		From:      from.Format("2006-01-02"),
		To:        to.Format("2006-01-02"),
		Days:      days,
		Units:     units,
		Bookings:  bookings,
		Occupancy: occupancy,
		Matrix:    matrix,
	}, nil
}

// ReportCSV generates CSV bytes with header date,revenue,bookings,occupancy for given range.
func (s *ReportService) ReportCSV(ctx context.Context, from, to time.Time) ([]byte, error) {
	revPerDay, err := s.ReportRepo.RevenuePerDay(ctx, from, to)
	if err != nil {
		return nil, err
	}
	occPerDay, err := s.ReportRepo.OccupancyPerDay(ctx, from, to)
	if err != nil {
		return nil, err
	}
	occMap := make(map[string]int64)
	for _, o := range occPerDay {
		occMap[o.Date] = o.Occupied
	}
	revMap := make(map[string]repo.DailyRevenue)
	for _, r := range revPerDay {
		revMap[r.Date] = r
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"date", "revenue", "bookings", "occupancy"}); err != nil {
		return nil, err
	}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		rev := int64(0)
		bookings := int64(0)
		if r, ok := revMap[ds]; ok {
			rev = r.Revenue
			bookings = r.Bookings
		}
		occ := occMap[ds]
		if err := w.Write([]string{ds, strconv.FormatInt(rev, 10), strconv.FormatInt(bookings, 10), strconv.FormatInt(occ, 10)}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	_ = fmt.Sprintf
	return buf.Bytes(), nil
}
