# 12 — Phase 4: Reviews, Reports & Dashboard Polish (COMPLETE 2026-10-08)

> Reviews (one per checked_out), reports (occupancy/revenue), and beautiful WarmAura dashboard with stats cards + charts. Live verified.

## Summary

| Item | Status | Evidence |
|------|--------|----------|
| **Reviews migration** | ✅ 002_reviews.sql | reviews table (booking_id UNIQUE, rating 1-5, comment) + room_types avg_rating/review_count |
| **Reviews API** | ✅ POST /api/reviews + GET /api/reviews?room_type_id + GET /api/room-types (with avg_rating) | Only checked_out, one per booking (UNIQUE), rating 1-5, ownership 403, live: booking 35 → review 5 stars → avg_rating 5 |
| **Reports API** | ✅ GET /api/admin/reports/summary|revenue|occupancy | Occupancy rate, revenue, by_status, by_room_type, per-day, RBAC owner/manager, date range validation |
| **Frontend Polish** | ✅ AdminView stats + charts | 4 stats cards (Lucide), DatePicker range, 3 Chart.js charts WarmAura #8B5A2B, HomeView stars + review form, MyBookings, build 2044→354KB OK |
| **Builds** | ✅ go vet 0, go test ok, npm build 392ms | service+handler tests pass |
| **Live Flow** | ✅ create→upload→verify→checkin→checkout→review | Booking 35: 700k → waiting → verified → checked_in unit 2 → checked_out → review 5 “Amazing stay!” |

## Reviews

**Migration `002_reviews.sql`:**
```sql
CREATE TABLE reviews (id BIGSERIAL PK, booking_id UNIQUE FK bookings, user_id FK users, room_type_id FK room_types, rating 1-5, comment, created_at);
ALTER TABLE room_types ADD avg_rating DOUBLE PRECISION DEFAULT 0, review_count INT DEFAULT 0;
```

**Repo `review.go`:** Create, FindByBookingID, ListByRoomType, UpdateRoomTypeStats (AVG + COUNT), GetAverageByRoomType — $1 safe.

**Service `review.go`:** CreateReview validates: rating 1-5, comment ≤500, booking exists, booking.user_id == userID else 403, booking.status == checked_out else 409, one per booking (UNIQUE constraint → 409), then INSERT + UPDATE room_types avg_rating/review_count.

**Handlers `review.go`:**
- `POST /api/reviews` → Auth, body {booking_id, rating, comment} → 201 {id, rating 5}
- `GET /api/reviews?room_type_id` → public, returns list
- `GET /api/room-types` → updated to JOIN avg_rating/review_count

**Tests `review_test.go`:** rating bounds, checked_out only, forbidden, one-per-booking.

**Live:**
```bash
POST /api/reviews {"booking_id":35,"rating":5,"comment":"Amazing stay!"} → 201 {"id":1}
GET /api/reviews?room_type_id=1 → [{"rating":5}]
GET /api/room-types → Standard avg_rating 5 review_count 1
```

## Reports

**Service `report.go`:** GetOccupancyRate (occupied unit-days / total*18 *100), GetRevenue SUM total_price WHERE status verified/checked_in/checked_out and check_in between from/to, GetBookingsStats count per status, revenue per room_type, per-day arrays. Default last 30 days if no from/to, validates from>to 400, invalid date 400.

**Handlers `report.go`:**
- `GET /api/admin/reports/summary?from&to` → RBAC owner/manager → {occupancy_rate, total_revenue, total_bookings, by_status, by_room_type, revenue_per_day, occupancy_per_day}
- `GET /api/admin/reports/revenue`, `occupancy` → similar

**Live:**
```bash
GET /api/admin/reports/summary?from=2026-10-01&to=2026-10-31 (owner) → total_revenue 700k, occupancy_rate 5.5%, by_status checked_out 1
GET with customer token → 403
GET from>to → 400
```

## Frontend Polish

**AdminView.vue:**
- 4 stats cards grid 1/2/4: Total Bookings (CalendarDays #8B5A2B), Revenue (Wallet #C9A86A), Occupancy (TrendingUp #1A3A4A), Available Units (Bed cream)
- DatePicker range → fetches summary?from&to
- 3 Chart.js charts via PrimeVue Chart: Line revenue per day, Bar bookings by room type, Doughnut occupancy (WarmAura colors #8B5A2B, #C9A86A, #1A3A4A, #FDF6EC)
- Existing verify/checkin/checkout + room-units table kept

**HomeView.vue + Reviews:**
- Per room_type: avg_rating Stars (PrimeVue Rating) + review_count
- List reviews per type cards
- Review form: Rating + Textarea, only shown if user has checked_out booking without review → POST /api/reviews

**Build:** `npm run build` 354KB AdminView, HomeView 25KB, rating 188KB, ✓ 392ms.

## Verification

```bash
go vet ./...          # 0
go test ./...         # ok service+handler 1.29s
npm run build         # 2044→354KB OK
curl POST /api/reviews booking 35 checked_out → 201 rating 5
curl GET /api/reviews?1 → 1 review
curl GET /api/room-types → Standard 5.0/1
curl GET /api/admin/reports/summary → revenue 700k
```

## DoD Phase 4

- [x] Reviews one per checked_out, avg_rating, tests
- [x] Reports occupancy/revenue/by_status/by_type, RBAC, tests
- [x] Dashboard 4 cards + 3 charts WarmAura #8B5A2B + HomeView stars + build OK
- [ ] Next: final polish (invoice PDF, email notif, seed voucher, prod hardening)

## How to Run Phase 4

```bash
cd backend && go run ./cmd/server  # :8080
# Review after checkout
curl -X POST :8080/api/reviews -H "Bearer $CUSTOMER" -d '{"booking_id":35,"rating":5,"comment":"Amazing!"}'
# Reports
curl :8080/api/admin/reports/summary?from=2026-10-01&to=2026-10-31 -H "Bearer $OWNER"
cd frontend && npm run dev  # :5173 AdminView stats + charts
```

> All SRP, $1 safe, WarmAura #8B5A2B, checked_out only, UNIQUE, RBAC.
