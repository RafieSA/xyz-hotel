# 09 — Phase 1 Scaffolding (LOCKED 2026-10-08 — 100% COMPLETE)

> One full iteration: GitHub repo + Go Fiber backend + Vue PrimeVue Aura frontend + build verification.

## Summary

| Item | Status | Detail |
|------|--------|--------|
| **GitHub Repo** | ✅ https://github.com/RafieSA/xyz-hotel | public, remote `origin` configured |
| **Backend** | ✅ `backend/` | Go 1.26.3 + Fiber v2 + pgx/sqlx + migrate + validator + jwt + godotenv, `go vet` & `go build` OK |
| **Frontend** | ✅ `frontend/` | Vue 3 + Vite 8 + Tailwind 3.4 + PrimeVue 4.5 Aura WarmAura `#8B5A2B` + Lucide + Router + Pinia + Zod, `npm run build` OK (1969 modules) |
| **Build Verification** | ✅ | `go vet ./...` exit 0, `vite build` 1969 modules, no errors |
| **Git Ops** | ⏸️ Paused | Per instruction "no git ops at all" — files created but not yet `git add/commit/push` |

## Repository Structure After Phase 1

```
xyz-hotel/
├── backend/
│   ├── cmd/server/main.go          # Fiber :8080, CORS 5173, /health, /api/health, slog
│   ├── internal/
│   │   ├── handler/health.go + booking.go (stub)
│   │   ├── middleware/auth.go (JWT) + rbac.go (4 roles)
│   │   ├── model/user.go, room.go, booking.go, voucher.go, audit.go
│   │   ├── repo/db.go + room.go + booking.go + user.go
│   │   └── service/availability.go (FOR UPDATE placeholder)
│   ├── migrations/001_init.sql     # users, room_types, room_units, bookings, vouchers, audit_logs
│   ├── .env.example                # DATABASE_URL, JWT_SECRET, PORT
│   ├── Makefile                    # run, migrate, tidy
│   ├── go.mod (module xyz-hotel/backend)
│   └── README.md
├── frontend/
│   ├── src/
│   │   ├── main.js                 # createApp + Pinia + Router + PrimeVue WarmAura
│   │   ├── theme/aura.js           # WarmAura preset #8B5A2B
│   │   ├── router/index.js         # / (Home), /admin, /login
│   │   ├── stores/auth.js          # Pinia auth
│   │   ├── views/HomeView.vue      # Hero + search box + 4 room Cards (PrimeVue Card + Lucide) elegant
│   │   ├── views/AdminView.vue     # DataTable + Tag + Chart placeholder
│   │   └── assets/
│   ├── tailwind.config.js          # colors primary #8B5A2B, cream, teal, gold
│   ├── vite.config.js
│   ├── index.html
│   └── dist/ (build output)
├── docs/
│   ├── 00-goal.md ... 08-dependencies.md (Phase 0)
│   └── 09-fase1.md (this file)
├── AGENTS.md
└── README.md
```

## Backend Details

**go.mod:**
```
module xyz-hotel/backend
go 1.26
fiber v2.52.15, pgx v5.11.0, sqlx, migrate v4.20.1, validator v10, jwt v5, godotenv
```

**Phase 1 Endpoints:**
- `GET /health` → `{"status":"ok","service":"xyz-hotel"}`
- `GET /api/health` → same
- CORS `http://localhost:5173` (Vite frontend)

**Migrations 001_init.sql:**
- `users` (id, email unique, password_hash, role enum owner/manager/receptionist/customer, index on email)
- `room_types` (4 types to be seeded: Standard/Deluxe/Family/Suite)
- `room_units` (18 units: STD-101..108, DLX-201..205, FAM-301..303, STE-401..402, status available/occupied/dirty/maintenance)
- `bookings` (room_type_id, user_id, check_in/out, status pending_payment→checked_out, total_price snapshot, proof_url, index on check_in/out/type/status)
- `vouchers` + `audit_logs`

**How to Run:**
```bash
cd backend
cp .env.example .env  # fill in DATABASE_URL=postgres://...
go mod tidy
go run ./cmd/server   # :8080
# or
make run
```

## Frontend Details

**Stack:**
- Vue 3.5 + Vite 8 + vue-router 5 + pinia 4
- Tailwind 3.4 (content src/**/*.{vue,js}, extend colors #8B5A2B)
- PrimeVue 4.5 + @primevue/themes Aura WarmAura
- lucide-vue-next, axios, zod, vee-validate

**Routes:**
- `/` → HomeView (hotel hero image + overlay #1A3A4A 40% + cream rounded-2xl search box + 4 room Cards)
- `/admin` → AdminView (bookings DataTable + status-colored Tag + Chart placeholder)
- `/login` → LoginView

**WarmAura Preset:**
```js
// src/theme/aura.js
definePreset(Aura, { semantic: { primary: {500:'#8B5A2B', 600:'#6F4620'}, colorScheme:{light:{primary:{color:'{primary.500}'}}}}})
```

**How to Run:**
```bash
cd frontend
npm install
npm run dev    # http://localhost:5173
npm run build  # dist/
```

## Phase 1 Verification (Factual)

```bash
cd backend && go vet ./...   # exit 0
cd frontend && npm run build # 1969 modules, built in 231ms, no errors
```

## Remaining (Phase 2)

- Connect Postgres DB & run migration `001_init.sql`
- Real JWT auth (register/login) + RBAC 4 roles
- Real availability `FOR UPDATE` + booking CRUD
- Proof upload + verification + 12-hour expiry ticker
- Seed 18 room units
- API docs

## Phase 1 Definition of Done

- [x] Public GitHub repo created
- [x] Backend hello world + SRP structure + go.mod + build OK
- [x] Frontend hello world + WarmAura #8B5A2B + build OK
- [x] This documentation
- [ ] Git add/commit/push paused (no git ops)

> Next: Phase 2 — DB migration + auth + availability (no git ops until you request it).
