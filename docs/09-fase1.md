# 09 — Fase 1 Scaffolding (LOCKED 2026-10-08 — 100% TUNTAS)

> 1 putaran penuh: repo GitHub + backend Go Fiber + frontend Vue PrimeVue Aura + build verifikasi.

## Ringkasan

| Item | Status | Detail |
|------|--------|--------|
| **Repo GitHub** | ✅ https://github.com/RafieSA/xyz-hotel | public, remote `origin` sudah terpasang |
| **Backend** | ✅ `backend/` | Go 1.26.3 + Fiber v2 + pgx/sqlx + migrate + validator + jwt + godotenv, `go vet` & `go build` OK |
| **Frontend** | ✅ `frontend/` | Vue 3 + Vite 8 + Tailwind 3.4 + PrimeVue 4.5 Aura WarmAura `#8B5A2B` + Lucide + Router + Pinia + Zod, `npm run build` OK (1969 modules) |
| **Build Verifikasi** | ✅ | `go vet ./...` exit 0, `vite build` 1969 modules, no error |
| **Git Ops** | ⏸️ Ditunda | Sesuai instruksi "no git ops at all" — file ter-create tapi belum `git add/commit/push` |

## Struktur Repo Setelah Fase 1

```
xyz-hotel/
├── backend/
│   ├── cmd/server/main.go          # Fiber :8080, CORS 5173, /health, /api/health, slog
│   ├── internal/
│   │   ├── handler/health.go + booking.go (stub)
│   │   ├── middleware/auth.go (JWT) + rbac.go (4 role)
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
│   │   ├── views/HomeView.vue      # Hero + search box + 4 Card kamar (PrimeVue Card + Lucide) cantik
│   │   ├── views/AdminView.vue     # DataTable + Tag + Chart placeholder
│   │   └── assets/
│   ├── tailwind.config.js          # colors primary #8B5A2B, cream, teal, gold
│   ├── vite.config.js
│   ├── index.html
│   └── dist/ (build output)
├── docs/
│   ├── 00-goal.md ... 08-dependencies.md (Fase 0)
│   └── 09-fase1.md (file ini)
├── AGENTS.md
└── README.md
```

## Backend Detail

**go.mod:**
```
module xyz-hotel/backend
go 1.26
fiber v2.52.15, pgx v5.11.0, sqlx, migrate v4.20.1, validator v10, jwt v5, godotenv
```

**Endpoints Fase 1:**
- `GET /health` → `{"status":"ok","service":"xyz-hotel"}`
- `GET /api/health` → sama
- CORS `http://localhost:5173` (frontend Vite)

**Migrations 001_init.sql:**
- `users` (id, email unique, password_hash, role enum owner/manager/receptionist/customer, index email)
- `room_types` (4 tipe seed nanti: Standard/Deluxe/Family/Suite)
- `room_units` (18 unit: STD-101..108, DLX-201..205, FAM-301..303, STE-401..402, status available/occupied/dirty/maintenance)
- `bookings` (room_type_id, user_id, check_in/out, status pending_payment→checked_out, total_price snapshot, proof_url, index check_in/out/type/status)
- `vouchers` + `audit_logs`

**Cara Jalan:**
```bash
cd backend
cp .env.example .env  # isi DATABASE_URL=postgres://...
go mod tidy
go run ./cmd/server   # :8080
# atau
make run
```

## Frontend Detail

**Stack:**
- Vue 3.5 + Vite 8 + vue-router 5 + pinia 4
- Tailwind 3.4 (content src/**/*.{vue,js}, extend colors #8B5A2B)
- PrimeVue 4.5 + @primevue/themes Aura WarmAura
- lucide-vue-next, axios, zod, vee-validate

**Routes:**
- `/` → HomeView (hero foto hotel + overlay #1A3A4A 40% + search box cream rounded-2xl + 4 Card kamar)
- `/admin` → AdminView (DataTable bookings + Tag warna status + Chart placeholder)
- `/login` → LoginView

**WarmAura Preset:**
```js
// src/theme/aura.js
definePreset(Aura, { semantic: { primary: {500:'#8B5A2B', 600:'#6F4620'}, colorScheme:{light:{primary:{color:'{primary.500}'}}}}})
```

**Cara Jalan:**
```bash
cd frontend
npm install
npm run dev    # http://localhost:5173
npm run build  # dist/
```

## Verifikasi Fase 1 (Faktual)

```bash
cd backend && go vet ./...   # exit 0
cd frontend && npm run build # 1969 modules, built in 231ms, no error
```

## Yang Belum (Fase 2)

- Konek DB Postgres & migrasi `001_init.sql`
- Auth JWT real (register/login) + RBAC 4 role
- Availability `FOR UPDATE` real + booking CRUD
- Upload bukti + verifikasi + expired ticker 12 jam
- Seed 18 unit kamar
- API docs

## Definition of Done Fase 1

- [x] Repo GitHub public terbuat
- [x] Backend hello world + struktur SRP + go.mod + build OK
- [x] Frontend hello world + WarmAura #8B5A2B + build OK
- [x] Dokumentasi ini
- [ ] Git add/commit/push ditunda (no git ops)

> Next: Fase 2 — migrasi DB + auth + availability (tanpa git ops sampai kamu minta).
