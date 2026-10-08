# xyz-hotel

Aplikasi booking hotel berbasis website — **Frontoffice (pelanggan)** + **Backoffice (manajemen hotel)**. Satu hotel.

> Status: **FASE 0 — Grill & Brainstorming** (belum coding, lagi matangkan ide & dokumentasi)

## Posisi Saat Ini
- Folder: `/Users/rafiesafarazaribowo/Projects/xyz-hotel`
- Branch: `main` (local only, belum push ke GitHub)
- Device: MacBook Air M4, local Postgres 18.4 / MySQL 9.6, PHP 8.2, Node 26
- Aturan keras: **NO Next.js**

## Struktur Rencana (akan jadi monorepo)
```
xyz-hotel/
├── app/          # Frontend website (pilihan stack masih di-grill)
├── backend/      # API + logic booking
├── docs/         # Dokumentasi Fase 0 (sudah ada 6 file)
├── AGENTS.md     # Aturan kerja agent
└── README.md
```

## Dokumen Fase 0
| File | Isi |
|------|-----|
| `docs/00-goal.md` | Tujuan, visi, batasan project |
| `docs/01-brainstorming-log.md` | Log percakapan sesi 2026-10-08 |
| `docs/02-prd.md` | PRD level lengkap |
| `docs/03-tech-stack-decision.md` | Keputusan tech stack (tanpa Next.js) + tradeoff |
| `docs/04-roles-permissions.md` | 4 role + matriks izin |
| `docs/05-booking-flow-and-edge-cases.md` | Alur booking + edge cases ekstrem |
| `docs/06-context-and-constraints.md` | Konteks portfolio, local-first, security |

## Cara Jalan (nanti setelah stack final)
```bash
# Belum ada — akan diisi setelah grill selesai
```

## Definition of Done Fase 0
- [x] Folder + git init lokal
- [x] 6 file docs terisi
- [ ] Tech stack final disepakati (tanpa Next.js)
- [ ] 4 role + flow booking disepakati
- [ ] Siap lanjut ke Fase 1 (scaffolding)

## Next Step
Jawab grill Q1 di `docs/03-tech-stack-decision.md` — pilih: Laravel + Inertia Vue / Nuxt / SvelteKit / Dart Frog.
