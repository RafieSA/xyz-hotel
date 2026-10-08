# AI Agent Rules (baca setiap sesi)

Project: xyz-hotel. Aplikasi booking hotel berbasis website. Satu hotel. Monorepo `app/ + backend/ + docs/`. Branch `main`, kerja di local branch, no worktrees. Local only (MacBook Air M4), tanpa Docker, Postgres/MySQL lokal.

> HARD RULE (user, 2026-10-08): BIG NO Next.js. Jangan pernah rekomendasikan Next.js lagi. Fase 0 = grill, brainstorming, discuss, dokumentasi dulu. No coding sebelum flow & tech stack final.

## Locked decisions (Fase 0)
- Satu hotel dulu (scalable ke multi-hotel nanti, tapi DB siap).
- Level Lengkap: searching, booking, availability real-time, review, voucher, laporan.
- 4 role (detail di `docs/04-roles-permissions.md`).
- Pembayaran manual transfer + upload bukti + verifikasi admin (no payment gateway di v1).
- Availability calendar WAJIB dengan handling race condition & ACID.
- Portfolio pribadi, no deadline, no client, no bos — bebas sekreatif mungkin, tapi tetap Clean Code.
- Local-first: Postgres 18.4 / MySQL 9.6, PHP 8.2, Node 26, Bun 1.4, Git 2.53 tersedia. Docker TIDAK ADA.

## How to work with Rafie
- Interaktif satu-satu, bahasa Indonesia membumi, selalu pakai contoh + tradeoff + edge case + visual (tabel/diagram).
- No shortcuts, no workaround, no happy-path-only. Setiap edge case harus di-handle.
- Jangan handwritten from scratch kalau sudah ada package/template resmi — manfaatkan yang ada.
- Clean Code, SRP, strong error handling + logging, testable, dokumentasi API jelas. YAGNI/DRY/KISS/SOLID/ACID.
- Block: IDOR, BOLA, BFLA, Broken Auth, SQLi (wajib parameterized query), XSS, CSRF, Sensitive Data Exposure, Security Misconfig, SSRF. No spaghetti, no future debt.
- Ground setiap jawaban di `docs/00..06` + kode riil. Jangan asumsi; audit dulu.
- Tanya saya lebih detail dan mendalam — kalau ragu, tanya, jangan tebak.

## Verification commands (akan dilengkapi setelah stack final)
- Backend: `composer test` / `php artisan test` atau `bun test` tergantung stack
- Frontend: `npm run lint` / `npm run build`
- DB: migrasi terurut `001_*.sql` atau `migrate:fresh`
