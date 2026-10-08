# AI Agent Rules (read every session)

Project: xyz-hotel. Website-based hotel booking application. Single hotel. Monorepo `app/ + backend/ + docs/`. Branch `main`, work on local branches, no worktrees. Local only (MacBook Air M4), no Docker, local Postgres/MySQL.

> HARD RULE (user, 2026-10-08): BIG NO to Next.js. Never recommend Next.js again. Phase 0 = grilling, brainstorming, discussion, and documentation first. No coding before flow & tech stack are finalized.
## Locked Decisions (Phase 0 - LOCKED 2026-10-08)
- Stack: **Backend Go 1.26.3 (Fiber + pgx/sqlx + migrate) + Frontend Vue 3 + Vite + Tailwind + PrimeVue Aura (WarmAura #8B5A2B) + Lucide + Postgres 18.4**. BIG NO to Next.js & shadcn (overused).
- Styling: **Tailwind CSS** (priority #1 LOOKS GREAT, not the engine).
- UI: **PrimeVue Aura custom** + Lucide, the most elegant choice for a hotel (premium DataTable, DatePicker, Card).
- Single hotel for now (scalable to multi-hotel later, but DB is ready).
- Full feature set: search, booking, real-time availability, reviews, vouchers, reporting.
- 4 roles: owner, manager, receptionist, customer (details in `docs/04-roles-permissions.md`).
- Manual payment via bank transfer + proof upload + admin verification (no payment gateway in v1).
- Availability calendar REQUIRED with race-condition & ACID handling (`FOR UPDATE`).
- 4 room types locked (Standard 8, Deluxe 5, Family 3, Suite 2 = 18 units, to be seeded).
- 12-hour expiry (`pending_payment` → `expired` if `created_at < now-12h`), check-in 14:00 / check-out 12:00.
- Design token locked `#8B5A2B` (Warm Brown) in `docs/07-design.md`, dependencies locked in `docs/08-dependencies.md`.
- Personal portfolio, no deadline, no client, no manager. Be as creative as you want, but keep it Clean Code.
- Local-first: Postgres 18.4, Go 1.26.3, Node 26, Bun 1.4, Git 2.53 available. Docker NOT AVAILABLE.
## How to Work with Rafie
- Interactive one-on-one, approachable and grounded English, always include examples + trade-offs + edge cases + visuals (tables/diagrams).
- No shortcuts, no workarounds, no happy-path-only. Every edge case must be handled.
- Do not handwrite from scratch if an official package or template already exists. Leverage what is available.
- Clean Code, SRP, strong error handling + logging, testable, clear API documentation. YAGNI/DRY/KISS/SOLID/ACID.
- Block: IDOR, BOLA, BFLA, Broken Auth, SQLi (parameterized queries required), XSS, CSRF, Sensitive Data Exposure, Security Misconfiguration, SSRF. No spaghetti, no future debt.
- Ground every answer in `docs/00..06` + real code. Don't assume; audit first.
- Ask for more detail and depth. When in doubt, ask, do not guess.

## Verification Commands (to be completed after final stack)
- Backend: `composer test` / `php artisan test` or `bun test` depending on stack
- Frontend: `npm run lint` / `npm run build`
- DB: ordered migrations `001_*.sql` or `migrate:fresh`
