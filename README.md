# xyz-hotel

Aplikasi booking hotel berbasis website — **Frontoffice (pelanggan)** + **Backoffice (manajemen hotel)**. Satu hotel.

> Status: **FASE 1 — SCAFFOLDING 100% TUNTAS 2026-10-08** — Repo https://github.com/RafieSA/xyz-hotel + Backend Go Fiber + Frontend Vue PrimeVue Aura

## Stack Final (Locked)
| Lapisan | Teknologi |
|---------|-----------|
| Backend | Go 1.26.3 + Fiber + pgx/sqlx + golang-migrate + validator + jwt |
| Frontend | Vue 3 + Vite + Tailwind + PrimeVue Aura (WarmAura #8B5A2B) + Lucide + Router + Pinia |
| DB | Postgres 18.4 (local, tanpa Docker) |
| Auth | JWT 15m + refresh + bcrypt + RBAC 4 role |

## Posisi Saat Ini
- Repo: https://github.com/RafieSA/xyz-hotel (public, origin terpasang)
- Folder: `/Users/rafiesafarazaribowo/Projects/xyz-hotel`
- Branch: `main` (local, belum push — no git ops)
- Backend: `backend/` — Fiber :8080, `go vet` OK
- Frontend: `frontend/` — Vite :5173, `npm run build` 1969 modules OK
- Device: MacBook Air M4, Go 1.26.3, Node 26, Postgres 18.4
- Aturan: **NO Next.js — Go + Vue + PrimeVue Aura cantik**

## Struktur
```
xyz-hotel/
├── backend/        # Go Fiber
├── frontend/       # Vue 3 + Tailwind + PrimeVue Aura WarmAura
├── docs/           # 00..09 (10 file)
├── AGENTS.md
└── README.md
```

## Dokumen (10 file)
| File | Isi |
|------|-----|
| `docs/00-goal.md` | Tujuan 1 hotel |
| `docs/01-brainstorming-log.md` | Log sesi 2026-10-08 |
| `docs/02-prd.md` | PRD level lengkap |
| `docs/03-tech-stack-decision.md` | Go+Vue tanpa Next.js |
| `docs/04-roles-permissions.md` | 4 role |
| `docs/05-booking-flow...md` | Flow + 12 edge cases, expired 12 jam |
| `docs/06-context...md` | Portfolio, 18 unit seed |
| `docs/07-design.md` | Design `#8B5A2B` cantik |
| `docs/08-dependencies.md` | Tailwind+PrimeVue Aura+Lucide |
| `docs/09-fase1.md` | Fase 1 scaffolding 100% |
## Cara Jalan Fase 1 (Terverifikasi)
```bash
# Backend
cd backend && go run ./cmd/server  # :8080 /health
# Frontend
cd frontend && npm run dev          # :5173
cd frontend && npm run build        # 1969 modules OK
```

## Definition of Done
- [x] Fase 0 — 9 file docs (00-08) locked ✅
- [x] Fase 1 — Repo https://github.com/RafieSA/xyz-hotel + backend Fiber + frontend PrimeVue Aura + build OK ✅ (no git ops)
- [ ] Fase 2 — Migrasi DB + auth JWT + availability FOR UPDATE

## Next Step
Fase 2 — konek Postgres, migrasi 001, seed 18 unit, auth 4 role
