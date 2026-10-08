# xyz-hotel: Book a Room in 90 Seconds. Run the Hotel From One Dashboard.

**One hotel, two sides that work together.** Customers search real dates, see live availability, book, and upload payment proof in under two minutes. Staff verify bookings, assign rooms, update status, and track revenue without a spreadsheet.

> Status: **PHASE 1 SCAFFOLDING 100 PERCENT COMPLETE 2026-10-08**. Repo https://github.com/RafieSA/xyz-hotel. Go Fiber backend plus Vue PrimeVue Aura frontend. Build verified.

## Final Stack (Locked)
| Layer | Technology |
|---------|-----------|
| Backend | Go 1.26.3 + Fiber + pgx/sqlx + golang-migrate + validator + jwt |
| Frontend | Vue 3 + Vite + Tailwind + PrimeVue Aura (WarmAura #8B5A2B) + Lucide + Router + Pinia |
| DB | Postgres 18.4 (local, no Docker) |
| Auth | JWT 15m + refresh + bcrypt + RBAC 4 roles |

## Current Status
- Repo: https://github.com/RafieSA/xyz-hotel (public, origin configured)
- Directory: `/Users/rafiesafarazaribowo/Projects/xyz-hotel`
- Branch: `main` (local, not yet pushed, no git ops)
- Backend: `backend/` on Fiber :8080, `go vet` OK
- Frontend: `frontend/` on Vite :5173, `npm run build` 1969 modules OK
- Device: MacBook Air M4, Go 1.26.3, Node 26, Postgres 18.4
- Rule: **NO Next.js. Go + Vue + PrimeVue Aura elegant**

## Structure
```
xyz-hotel/
├── backend/ # Go Fiber
├── frontend/ # Vue 3 + Tailwind + PrimeVue Aura WarmAura
├── docs/ # 00..09 (10 files)
├── AGENTS.md
└── README.md
```

## Documentation (10 files)
| File | Contents |
|------|-----|
| `docs/00-goal.md` | Single-hotel goal |
| `docs/01-brainstorming-log.md` | Session log 2026-10-08 |
| `docs/02-prd.md` | Full PRD |
| `docs/03-tech-stack-decision.md` | Go+Vue without Next.js |
| `docs/04-roles-permissions.md` | 4 roles |
| `docs/05-booking-flow...md` | Flow + 12 edge cases, 12-hour expiry |
| `docs/06-context...md` | Portfolio, 18-unit seed |
| `docs/07-design.md` | Design `#8B5A2B` elegant |
| `docs/08-dependencies.md` | Tailwind+PrimeVue Aura+Lucide |
| `docs/09-fase1.md` | Phase 1 scaffolding 100% |
## How to Run Phase 1 (Verified)
```bash
# Backend
cd backend && go run ./cmd/server # :8080 /health
# Frontend
cd frontend && npm run dev # :5173
cd frontend && npm run build # 1969 modules OK
```

## Definition of Done
- [x] Phase 0: 9 docs files (00-08) locked ✅
- [x] Phase 1: Repo https://github.com/RafieSA/xyz-hotel + Fiber backend + PrimeVue Aura frontend + build OK ✅ (no git ops)
- [ ] Phase 2: DB migration + JWT auth + availability FOR UPDATE

## Next Step
Phase 2: connect Postgres, migration 001, seed 18 units, 4-role auth
