# Changelog — xyz-hotel

All notable changes to this project are documented here. Format: `YYYY-MM-DD — what changed — why`.

## [Unreleased] — Enrich Extra 1,2,3,4 (YOLO, local)

- 2026-10-09 — Add chat support (WA widget + in-app WS) — guest can ask, admin can reply realtime — enrich extra 1
- 2026-10-09 — Add dark mode premium (cream ↔ deep, gold pop, no flash) — beautiful at night, portfolio unique — enrich extra 2
- 2026-10-09 — Add multi-language EN/ID (vue-i18n 89 keys, toggle persist) — international portfolio — enrich extra 3
- 2026-10-09 — Add interactive map booking (Leaflet 4 pins #8B5A2B, popup Reserve) — spatial wow — enrich extra 4
- 2026-10-09 — Add docs 16-enrich-extra-grill.md + 17-enrich-extra.md (English) — document grill and build

## 2026-10-09 — Enrich A+B+C (YOLO)

- 2026-10-09 — Add gallery 5 photos per room type (carousel + lightbox, upload admin) — premium guest wow
- 2026-10-09 — Add Leaflet map at Ubud with nearby cards (Beach, Cafe, Spa) — location trust
- 2026-10-09 — Add loyalty points (10 per night, 100 = IDR 100k) with badge and redeem — business logic showcase
- 2026-10-09 — Add occupancy calendar (7 days x 18 units) with colors — ops excellence
- 2026-10-09 — Add housekeeping Kanban (4 columns drag) — admin anti-spreadsheet
- 2026-10-09 — Add add-ons (Breakfast, Transfer, Extra bed) with total calc — revenue boost
- 2026-10-09 — Add dynamic pricing weekend +20% (Fri/Sat) — pricing logic
- 2026-10-09 — Add realtime WebSocket admin (new booking toast) — proper engineering
- 2026-10-09 — Add full-text search ILIKE for room types — discoverability
- 2026-10-09 — Add room photo upload (5 per type, admin) — gallery source
- 2026-10-09 — Add export reports CSV — ops reporting
- 2026-10-09 — Add rate limit 60/min with 429 — hardening
- 2026-10-09 — Add migrations 005_gallery.sql, 006_addons.sql, 007_loyalty.sql — schema for enrich
- 2026-10-09 — Add docs 14-enrich-grill.md + 15-enrich.md (English) — document A+B+C

## 2026-10-08 — Phase 5 YOLO

- 2026-10-08 — Add invoice PDF premium WarmAura (header #8B5A2B, gold line, hierarchy) — after verified only, 409 otherwise
- 2026-10-08 — Add wishlist/favorites (heart #8B5A2B toggle, enriched list) — guest save
- 2026-10-08 — Add email log only (async to backend/logs/email.log, non-blocking) — notify without SMTP
- 2026-10-08 — Add CRUD for room types and room units (soft delete) — admin management
- 2026-10-08 — Add cancel booking (pending/waiting only, 409 otherwise) — guest flexibility
- 2026-10-08 — Add audit log view (admin DataTable) — traceability
- 2026-10-08 — Add migrations 003_wishlist.sql, 004_room_soft.sql — wishlist and soft delete
- 2026-10-08 — Add docs 13-fase5.md (English) — Phase 5 summary
- 2026-10-08 — Add gofpdf dependency for invoice — PDF generation

## 2026-10-08 — Copy Upgrade

- 2026-10-08 — Revise all copy to clear, specific, active voice, no em dashes, no exclamation — conversion and readability

## 2026-10-08 — Phase 4

- 2026-10-08 — Add reviews (one per checked_out, rating 1-5, avg_rating) — trust
- 2026-10-08 — Add reports (occupancy, revenue, by status/type, per day) — business insight
- 2026-10-08 — Polish dashboard (4 stats cards, 3 Chart.js charts WarmAura) — beautiful ops
- 2026-10-08 — Add migration 002_reviews.sql — reviews schema

## 2026-10-08 — Phase 3

- 2026-10-08 — Add voucher CRUD and validation (discount, quota, expiry, min_nights) — promo
- 2026-10-08 — Add proof upload (MIME magic, 5MB, IDOR, pending → waiting) — payment proof
- 2026-10-08 — Add verification (waiting → verified/rejected, RBAC owner/manager) — ops control
- 2026-10-08 — Add check-in/out with room unit assignment (FOR UPDATE, dirty/available) — stay flow
- 2026-10-08 — Add frontend wishlist upload and admin verify/checkin/checkout UI — complete ops UI

## 2026-10-08 — Phase 2

- 2026-10-08 — Create database xyz_hotel and apply 001_init.sql (6 tables) — foundation
- 2026-10-08 — Seed 4 room types (Standard, Deluxe, Family, Suite) and 18 units (STD-101..STE-402) — inventory
- 2026-10-08 — Seed 4 users (owner, manager, receptionist, customer) with bcrypt — auth baseline
- 2026-10-08 — Implement auth (register/login, JWT 15m + refresh 7d, bcrypt, RBAC 4 roles) — security
- 2026-10-08 — Implement availability (FOR UPDATE transactional booking, price snapshot, expiry ticker 12h) — ACID
- 2026-10-08 — Wire frontend auth and availability UI — customer flow

## 2026-10-08 — English Translation

- 2026-10-08 — Translate all documentation (AGENTS.md, README.md, docs 00-09) to English — international portfolio

## 2026-10-08 — Phase 1 Scaffolding

- 2026-10-08 — Create GitHub repo RafieSA/xyz-hotel (public) — remote origin
- 2026-10-08 — Scaffold backend Go 1.26 + Fiber v2 + pgx/sqlx + migrate + validator + jwt — hello world
- 2026-10-08 — Scaffold frontend Vue 3 + Vite 8 + Tailwind 3.4 + PrimeVue 4.5 Aura WarmAura #8B5A2B + Lucide — hello world
- 2026-10-08 — Verify builds (go vet 0, vite 1969 modules) — quality gate

## 2026-10-08 — Phase 0 Grill

- 2026-10-08 — Lock dependencies Tailwind + PrimeVue Aura WarmAura #8B5A2B + Lucide + Go Fiber + pgx/sqlx (best choice for beauty) — design decision
- 2026-10-08 — Lock tech stack Go 1.26.3 + Vue 3 + Postgres 18.4 (G1, no Next.js) — architecture decision
- 2026-10-08 — Lock grill Q3/Q5/Q6 (4 roles, 12h expiry, 14:00/12:00, 18 units seed) + create 07-design.md (WarmAura palette) — business rules
- 2026-10-08 — Initial grill: 6 docs (goal, brainstorming log, PRD, stack, roles, flow, context) — Phase 0 foundation
