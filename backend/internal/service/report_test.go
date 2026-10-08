package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"xyz-hotel/backend/internal/repo"
)

func reportTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://rafiesafarazaribowo@localhost:5432/xyz_hotel?sslmode=disable"
	}
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		t.Skipf("no db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("db ping failed: %v", err)
	}
	return db
}

func TestParseReportRange_Defaults(t *testing.T) {
	from, to, err := ParseReportRange("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if to.Sub(from).Hours()/24 != 30 {
		t.Errorf("expected 30 days default, got from %v to %v", from, to)
	}
}

func TestParseReportRange_InvalidDate(t *testing.T) {
	_, _, err := ParseReportRange("invalid", "2026-10-10")
	if err == nil {
		t.Error("expected error for invalid from")
	}
	_, _, err = ParseReportRange("2026-10-10", "not-a-date")
	if err == nil {
		t.Error("expected error for invalid to")
	}
}

func TestParseReportRange_FromAfterTo(t *testing.T) {
	_, _, err := ParseReportRange("2026-10-15", "2026-10-10")
	if err == nil {
		t.Error("expected error when from>to")
	}
	if err != nil && err.Error() != "from must be before or equal to to" {
		t.Errorf("unexpected msg: %v", err)
	}
}

func TestReportCalculation_RevenueAndOccupancy(t *testing.T) {
	db := reportTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// Clean test date bookings for deterministic test (FK cascade deletes reviews)
	_, _ = db.Exec(`DELETE FROM bookings WHERE check_in='2026-10-10'`)
	// Ensure 18 units exist
	var unitCount int
	if err := db.Get(&unitCount, `SELECT COUNT(*) FROM room_units WHERE deleted_at IS NULL`); err != nil {
		t.Fatalf("count units: %v", err)
	}
	if unitCount != 18 {
		t.Logf("warning: expected 18 units got %d", unitCount)
	}
	// Get a user and room_type
	var userID int64
	if err := db.Get(&userID, `SELECT id FROM users LIMIT 1`); err != nil {
		t.Fatalf("get user: %v", err)
	}
	var rtID int64
	if err := db.Get(&rtID, `SELECT id FROM room_types LIMIT 1`); err != nil {
		t.Fatalf("get room_type: %v", err)
	}

	// Insert 2 verified bookings 700k each, single night 2026-10-10 to 2026-10-11
	from := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 2; i++ {
		_, err := db.Exec(`INSERT INTO bookings (user_id, room_type_id, check_in, check_out, guests, total_price, status) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			userID, rtID, "2026-10-10", "2026-10-11", 2, 700000, "verified")
		if err != nil {
			t.Fatalf("insert booking: %v", err)
		}
	}
	// Insert one pending that should NOT count for revenue/occupancy
	_, _ = db.Exec(`INSERT INTO bookings (user_id, room_type_id, check_in, check_out, guests, total_price, status) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		userID, rtID, "2026-10-10", "2026-10-11", 1, 999999, "pending_payment")

	svc := NewReportService(db)

	// Revenue should be 1.4M
	rev, err := svc.GetRevenue(ctx, from, to)
	if err != nil {
		t.Fatalf("GetRevenue: %v", err)
	}
	if rev != 1400000 {
		t.Errorf("revenue got %d want 1400000", rev)
	}

	// Occupancy: 2 occupied unit-days /18 *100 = 11.111...
	occ, err := svc.GetOccupancyRate(ctx, from, to)
	if err != nil {
		t.Fatalf("GetOccupancyRate: %v", err)
	}
	expectedOcc := float64(2) / 18.0 * 100
	if occ < expectedOcc-0.01 || occ > expectedOcc+0.01 {
		t.Errorf("occupancy got %.4f want %.4f", occ, expectedOcc)
	}

	// Summary check
	sum, err := svc.GetBookingsStats(ctx, from, to)
	if err != nil {
		t.Fatalf("GetBookingsStats: %v", err)
	}
	if sum.TotalRevenue != 1400000 {
		t.Errorf("summary revenue %d", sum.TotalRevenue)
	}
	if sum.TotalBookings != 3 {
		t.Errorf("total bookings got %d want 3", sum.TotalBookings)
	}
	if sum.ByStatus["verified"] != 2 {
		t.Errorf("by_status verified %d", sum.ByStatus["verified"])
	}
	if sum.ByStatus["pending_payment"] != 1 {
		t.Errorf("by_status pending %d", sum.ByStatus["pending_payment"])
	}
	if len(sum.ByRoomType) == 0 {
		t.Error("by_room_type empty")
	} else {
		// revenue per room_type should be 1.4M for that type
		if sum.ByRoomType[0].Revenue != 1400000 {
			t.Errorf("by_room_type revenue %d", sum.ByRoomType[0].Revenue)
		}
	}

	// No data range returns 0
	emptyFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	emptyTo := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	rev2, _ := svc.GetRevenue(ctx, emptyFrom, emptyTo)
	if rev2 != 0 {
		t.Errorf("empty revenue got %d want 0", rev2)
	}
	occ2, _ := svc.GetOccupancyRate(ctx, emptyFrom, emptyTo)
	if occ2 != 0 {
		t.Errorf("empty occupancy got %f want 0", occ2)
	}

	// Cleanup
	_, _ = db.Exec(`DELETE FROM bookings WHERE check_in='2026-10-10'`)

	_ = repo.NewReportRepo // ensure import used
}
