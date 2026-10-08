# 10 — Phase 2: Database, Auth & Availability (COMPLETE 2026-10-08)

> One full round: DB migration + seed + JWT auth (4 roles) + availability with FOR UPDATE + booking create + frontend wiring. Verified with live curl and builds.

## Summary

| Item | Status | Evidence |
|------|--------|----------|
| **Database xyz_hotel** | ✅ Created | `psql -l` → xyz_hotel exists, Postgres 18.6 |
| **Migration 001_init.sql** | ✅ Applied | 6 tables: users, room_types, room_units, bookings, vouchers, audit_logs |
| **Seed** | ✅ 4 types + 18 units + 4 users | `SELECT count(*) → 4/18/4` via psql |
| **Auth Backend** | ✅ JWT 15m + refresh 7d + bcrypt + RBAC | `go vet 0`, `go test ok`, curl register/login/me/refresh + RBAC 403/200 |
| **Availability + Booking** | ✅ FOR UPDATE transactional | `GET /api/availability` → available 8, `POST /api/bookings` → pending_payment, 409 when full, ticker 12h |
| **Frontend Wiring** | ✅ Auth + availability UI | `npm run build` 2030 modules OK |
| **Verification** | ✅ Live smoke | `GET /health`, `GET /api/availability`, `POST /api/auth/login` all OK (see logs below) |

## Database

**Created:**
```sql
CREATE DATABASE xyz_hotel;
\psql postgres://rafiesafarazaribowo@localhost:5432/xyz_hotel -f backend/migrations/001_init.sql
-- 6 tables + 3 enums (user_role, room_unit_status, booking_status)
```

**Seed Data (idempotent):**

| room_types | price | units |
|------------|-------|-------|
| Standard | 350,000 | 8 (STD-101..108) cap 2 |
| Deluxe | 550,000 | 5 (DLX-201..205) cap 2 |
| Family | 850,000 | 3 (FAM-301..303) cap 4 |
| Suite | 1,250,000 | 2 (STE-401..402) cap 4 |

**Users (bcrypt $2a$10$):**
- owner@xyz-hotel.local / Owner123! (owner)
- manager@xyz-hotel.local / Manager123! (manager)
- receptionist@xyz-hotel.local / Receptionist123! (receptionist)
- customer@xyz-hotel.local / Customer123! (customer)

**Files:**
- `backend/seed/seed.go` (idempotent ON CONFLICT, slog)
- `backend/cmd/seed/main.go` (godotenv + sqlx pgx)
- `backend/.env` (DATABASE_URL, JWT_SECRET, PORT)

**Verification:**
```bash
psql xyz_hotel -c "SELECT count(*) FROM room_types"  # 4
psql xyz_hotel -c "SELECT count(*) FROM room_units"  # 18
psql xyz_hotel -c "SELECT count(*) FROM users"       # 4
```

## Auth Backend

**Endpoints:**
- `POST /api/auth/register` → 201 {access_token, refresh_token, user} — role forced `customer`, password min 8, email unique via $1
- `POST /api/auth/login` → 200 {access_token, refresh_token, user} — bcrypt Compare
- `POST /api/auth/refresh` → 200 {access_token, refresh_token} — validates typ=refresh
- `GET /api/auth/me` → 200 {user} — requires Bearer access token

**Security:**
- Password hash `bcrypt.DefaultCost`, never logged
- JWT HMAC via `golang-jwt/jwt/v5`, access 15m, refresh 7d, claims {uid, role, email, typ, sub, exp, iat}
- Middleware `auth.go`: 401 on missing/invalid/expired/refresh misuse, sets `c.Locals("user")`
- RBAC `RequireRole(...)`: 403 on BFLA, logs audit via slog
- SQLi safe via `$1,$2` placeholders (sqlx)
- Validation via `go-playground/validator` (email, required, min 8)

**Routing in `cmd/server/main.go`:**
- `/api/auth` public
- `/api/availability` public
- `/api/bookings` → Auth middleware
- `/api/admin/*` → Auth + RequireRole(owner,manager)

**Tests:**
- `internal/service/auth_test.go`: HashPassword/CheckPassword, GenerateTokens expiry 15m/7d
- `internal/service/availability_test.go`: IsOverlapping (10 cases), CalculateAvailable (6 cases), NightsBetween (6 cases)
- `go vet ./...` exit 0, `go test ./...` ok (1.239s)

**Live Curl Evidence (2026-10-08):**
```bash
curl -X POST :8081/api/auth/login -d '{"email":"customer@xyz-hotel.local","password":"Customer123!"}'
# → 200 {"data":{"access_token":"eyJ...","refresh_token":"eyJ...","user":{...}}}

curl :8081/api/auth/me -H "Authorization: Bearer <access>"
# → 200 {"data":{"email":"customer@xyz-hotel.local","role":"customer"}}

curl :8081/api/admin/health -H "Authorization: Bearer <customer>"
# → 403 {"error":"forbidden: insufficient role"}

curl :8081/api/admin/health -H "Authorization: Bearer <owner>"
# → 200 admin ok
```

## Availability & Booking

**Service `internal/service/availability.go`:**
- `CheckAvailability(roomTypeID, checkIn, checkOut)` → {total_units, occupied, available}
- `CreateBooking(userID, roomTypeID, checkIn, checkOut, guests)` → transactional:
  ```sql
  BEGIN;
  SELECT * FROM room_types WHERE id=$1 FOR UPDATE;
  SELECT id FROM bookings WHERE room_type_id=$1 AND status IN ('verified','checked_in') AND check_in < $2 AND check_out > $3 FOR UPDATE;
  INSERT INTO bookings (user_id, room_type_id, check_in, check_out, total_price, status) VALUES (...) -- price snapshot
  INSERT INTO audit_logs
  COMMIT;
  ```
- `IsOverlapping`, `CalculateAvailable`, `NightsBetween` helpers

**Repo `internal/repo/booking.go`:**
- `CountOverlapping`, `CountOverlappingTx (FOR UPDATE)`, `CreateTx`, `ListByUser`, `ListAll`, `ExpirePending (12h)`

**Handler `internal/handler/booking.go`:**
- `GET /api/availability?room_type_id&check_in&check_out` → public, 400 on bad dates, 404 on missing type
- `POST /api/bookings` → auth, validation, 409 when `available==0`
- `GET /api/bookings` → auth, IDOR-safe (customer sees own, admin sees all)

**Expiry Ticker in `main.go`:**
```go
go func() {
  ticker := time.NewTicker(5 * time.Minute)
  // initial run after 10s + every 5m:
  // UPDATE bookings SET status='expired' WHERE status='pending_payment' AND created_at < now() - interval '12 hours'
}()
```

**Live Curl Evidence:**
```bash
curl "http://localhost:8081/api/availability?room_type_id=1&check_in=2026-10-15&check_out=2026-10-17"
# → {"data":{"room_type_id":1,"total_units":8,"occupied":0,"available":8}}

curl -X POST :8081/api/bookings -H "Authorization: Bearer <customer>" -d '{"room_type_id":1,"check_in":"2026-10-20","check_out":"2026-10-22","guests":2}'
# → {"data":{"id":1,"total_price":700000,"status":"pending_payment"}}
# — audit log inserted, 8 verified → next booking 409 conflict
```

## Frontend Wiring

**New Files:**
- `frontend/src/api/client.js` → axios baseURL `http://localhost:8080`, Bearer interceptor from localStorage
- `frontend/src/stores/auth.js` → loginRequest/registerRequest/fetchMe/refresh, localStorage tokens
- `frontend/src/views/RegisterView.vue` → calls registerRequest, PrimeVue Card + Input + Button #8B5A2B
- `frontend/src/views/LoginView.vue` → calls loginRequest, stores JWT
- `frontend/src/views/HomeView.vue` → watches check_in/out, calls `GET /api/availability`, shows Tag `available/occupied/total_units`, `POST /api/bookings` with auth, toast on success/409
- `frontend/src/router/index.js` → added `/register`

**Build:**
```bash
npm run build # 2030 modules, ✓ built in 372ms
# dist/assets: HomeView 112KB, AdminView 380KB, index 194KB, all gzip OK
```

## Verification (Factual 2026-10-08)

```bash
cd backend && go vet ./...          # exit 0
cd backend && go test ./...         # ok xyz-hotel/backend/internal/service (cached)
cd backend && psql xyz_hotel -c "SELECT count(*) FROM room_types"  # 4
curl :8081/health                   # {"status":"ok","service":"xyz-hotel"}
curl :8081/api/availability?...     # {"available":8}
curl :8081/api/auth/login           # 200 access+refresh
cd frontend && npm run build        # 2030 modules OK
```

## Definition of Done Phase 2

- [x] DB xyz_hotel + 001_init.sql + seed 4/18/4
- [x] Auth JWT 15m/7d + bcrypt + RBAC 4 roles + tests
- [x] Availability FOR UPDATE + booking create + expiry ticker + IDOR safe
- [x] Frontend auth + availability wiring + build OK
- [ ] Next: voucher + proof upload + verification + check-in/out + admin UI polish (Phase 3)

## How to Run Phase 2

```bash
# Backend
cd backend
cp .env.example .env  # already .env with xyz_hotel
go run ./cmd/seed     # idempotent seed
go run ./cmd/server   # :8080

# Frontend
cd frontend
npm install
npm run dev           # :5173
```

> All code is SRP, slog logging, $1 safe, JWT 15m/refresh, FOR UPDATE ACID, WarmAura #8B5A2B, no shortcuts.
