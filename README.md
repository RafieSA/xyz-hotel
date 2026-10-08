# xyz-hotel

Aplikasi booking hotel berbasis website — **Frontoffice (pelanggan)** + **Backoffice (manajemen hotel)**. Satu hotel.

> Status: **FASE 0 — LOCKED 2026-10-08** — Stack: **Go 1.26.3 + Vue 3 + Tailwind + PrimeVue Aura + Postgres 18.4**

## Stack Final (Locked)
| Lapisan | Teknologi |
|---------|-----------|
| Backend | Go 1.26.3 + Fiber + pgx/sqlx + golang-migrate + validator + jwt |
| Frontend | Vue 3 + Vite + Tailwind + PrimeVue Aura (WarmAura #8B5A2B) + Lucide |
| DB | Postgres 18.4 (local, tanpa Docker) |
| Auth | JWT 15m + refresh + bcrypt + RBAC 4 role |

## Posisi Saat Ini
- Folder: `/Users/rafiesafarazaribowo/Projects/xyz-hotel`
- Branch: `main` (local only, belum push ke GitHub)
- Device: MacBook Air M4, Go 1.26.3, Node 26, Postgres 18.4
- Aturan keras: **NO Next.js — Backend Go + Frontend Vue**
├── docs/         # Dokumentasi Fase 0 (sudah ada 6 file)
├── AGENTS.md     # Aturan kerja agent
└── README.md
```

## Dokumen Fase 0 (9 file)
| File | Isi |
|------|-----|
| `docs/00-goal.md` | Tujuan, visi, batasan project |
| `docs/01-brainstorming-log.md` | Log percakapan sesi 2026-10-08 |
| `docs/02-prd.md` | PRD level lengkap |
| `docs/03-tech-stack-decision.md` | Keputusan tech stack Go+Vue (tanpa Next.js) |
| `docs/04-roles-permissions.md` | 4 role + matriks izin |
| `docs/05-booking-flow-and-edge-cases.md` | Alur booking + 12 edge cases |
| `docs/06-context-and-constraints.md` | Konteks portfolio, local, security |
| `docs/07-design.md` | Design guidelines `#8B5A2B` cantik profesional |
| `docs/08-dependencies.md` | Dependencies locked Tailwind+PrimeVue Aura+Lucide |
## Cara Jalan (Fase 1 nanti)
```bash
# Backend Go
cd backend && go run ./cmd/server
# Frontend Vue
cd frontend && npm install && npm run dev
```

## Definition of Done Fase 0 — ✅ SELESAI
- [x] Folder + git init lokal
- [x] 8 file docs terisi (00-07 + AGENTS + README)
- [x] Tech stack locked: **Go + Vue + Postgres** (tanpa Next.js)
- [x] 4 role + flow booking locked (expired 12 jam, 14:00/12:00, 18 unit seed)
- [x] Design token locked `#8B5A2B`
- [ ] Siap lanjut ke Fase 1 (scaffolding) + push GitHub

## Next Step
Buat repo GitHub `xyz-hotel` + scaffolding `backend/` (Go Fiber) + `frontend/` (Vue Vite)
