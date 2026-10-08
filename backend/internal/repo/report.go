package repo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// ReportRepo handles aggregated reporting queries with $1 placeholders.
type ReportRepo struct {
	DB *sqlx.DB
}

func NewReportRepo(db *sqlx.DB) *ReportRepo { return &ReportRepo{DB: db} }

// RoomTypeStat holds bookings + revenue per room type.
type RoomTypeStat struct {
	Name     string `db:"name" json:"name"`
	Bookings int64  `db:"bookings" json:"bookings"`
	Revenue  int64  `db:"revenue" json:"revenue"`
}

// DailyRevenue holds revenue aggregated per day.
type DailyRevenue struct {
	Date     string `db:"day" json:"date"`
	Revenue  int64  `db:"revenue" json:"revenue"`
	Bookings int64  `db:"bookings" json:"bookings"`
}

// DailyOccupancy holds occupied units per day.
type DailyOccupancy struct {
	Date     string  `db:"day" json:"date"`
	Occupied int64   `db:"occupied" json:"occupied"`
	Rate     float64 `json:"rate"`
}

var validRevenueStatuses = []string{"verified", "checked_in", "checked_out"}

// TotalUnits returns count of active physical room units.
func (r *ReportRepo) TotalUnits(ctx context.Context) (int, error) {
	var cnt int
	if err := r.DB.GetContext(ctx, &cnt, `SELECT COUNT(*) FROM room_units WHERE deleted_at IS NULL`); err != nil {
		return 0, err
	}
	return cnt, nil
}

// OccupiedDays returns sum of occupied unit-days overlapping [from, to).
// from inclusive, to exclusive (treated as $2::date exclusive bound for correct day math).
// Uses LEAST(check_out, $2) - GREATEST(check_in, $1).
func (r *ReportRepo) OccupiedDays(ctx context.Context, from, to time.Time) (int64, error) {
	var days int64
	// to is inclusive date from caller; we convert to exclusive by adding 1 day for LEAST bound
	toExcl := to.AddDate(0, 0, 1)
	err := r.DB.GetContext(ctx, &days, `
		SELECT COALESCE(SUM( LEAST(check_out::date, $2::date) - GREATEST(check_in::date, $1::date) ), 0)
		FROM bookings
		WHERE status IN ('verified','checked_in','checked_out')
		  AND check_out > $1::date
		  AND check_in < $2::date
	`, from.Format("2006-01-02"), toExcl.Format("2006-01-02"))
	if err != nil {
		return 0, err
	}
	if days < 0 {
		days = 0
	}
	return days, nil
}

// Revenue returns SUM total_price for bookings with check_in BETWEEN from AND to inclusive and status in valid set.
func (r *ReportRepo) Revenue(ctx context.Context, from, to time.Time) (int64, error) {
	var rev int64
	err := r.DB.GetContext(ctx, &rev, `
		SELECT COALESCE(SUM(total_price),0)
		FROM bookings
		WHERE status IN ('verified','checked_in','checked_out')
		  AND check_in >= $1::date
		  AND check_in <= $2::date
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	return rev, err
}

// TotalBookings returns COUNT of all bookings where check_in between from and to.
func (r *ReportRepo) TotalBookings(ctx context.Context, from, to time.Time) (int64, error) {
	var cnt int64
	err := r.DB.GetContext(ctx, &cnt, `
		SELECT COUNT(*)
		FROM bookings
		WHERE check_in >= $1::date
		  AND check_in <= $2::date
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	return cnt, err
}

// CountByStatus returns count per booking_status filtered by check_in range.
func (r *ReportRepo) CountByStatus(ctx context.Context, from, to time.Time) (map[string]int64, error) {
	rows, err := r.DB.QueryxContext(ctx, `
		SELECT status, COUNT(*) as cnt
		FROM bookings
		WHERE check_in >= $1::date AND check_in <= $2::date
		GROUP BY status
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]int64)
	for rows.Next() {
		var status string
		var cnt int64
		if err := rows.Scan(&status, &cnt); err != nil {
			return nil, err
		}
		m[status] = cnt
	}
	return m, nil
}

// StatsByRoomType returns bookings and revenue per room_type for the date range.
// Bookings counts all bookings in range; revenue filtered to valid statuses.
func (r *ReportRepo) StatsByRoomType(ctx context.Context, from, to time.Time) ([]RoomTypeStat, error) {
	var stats []RoomTypeStat
	err := r.DB.SelectContext(ctx, &stats, `
		SELECT rt.name as name,
		       COUNT(b.id) as bookings,
		       COALESCE(SUM(CASE WHEN b.status IN ('verified','checked_in','checked_out') THEN b.total_price ELSE 0 END),0) as revenue
		FROM bookings b
		JOIN room_types rt ON rt.id = b.room_type_id
		WHERE b.check_in >= $1::date AND b.check_in <= $2::date
		GROUP BY rt.name
		ORDER BY rt.name
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if stats == nil {
		stats = []RoomTypeStat{}
	}
	return stats, nil
}

// RevenuePerDay returns revenue and bookings count per day.
func (r *ReportRepo) RevenuePerDay(ctx context.Context, from, to time.Time) ([]DailyRevenue, error) {
	var out []DailyRevenue
	err := r.DB.SelectContext(ctx, &out, `
		SELECT check_in::date::text as day,
		       COALESCE(SUM(CASE WHEN status IN ('verified','checked_in','checked_out') THEN total_price ELSE 0 END),0) as revenue,
		       COUNT(*) as bookings
		FROM bookings
		WHERE check_in >= $1::date AND check_in <= $2::date
		GROUP BY check_in::date
		ORDER BY day
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []DailyRevenue{}
	}
	return out, nil
}

// OccupancyPerDay returns occupied count per day using generate_series.
func (r *ReportRepo) OccupancyPerDay(ctx context.Context, from, to time.Time) ([]DailyOccupancy, error) {
	var out []DailyOccupancy
	err := r.DB.SelectContext(ctx, &out, `
		SELECT gs::date::text as day,
		       COUNT(b.id) as occupied
		FROM generate_series($1::date, $2::date, '1 day'::interval) gs
		LEFT JOIN bookings b ON b.status IN ('verified','checked_in','checked_out')
		  AND b.check_in <= gs::date
		  AND b.check_out > gs::date
		GROUP BY gs::date
		ORDER BY gs::date
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []DailyOccupancy{}
	}
	return out, nil
}
