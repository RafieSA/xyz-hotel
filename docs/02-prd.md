# 02 — PRD (Product Requirements Document) — Complete Level

> Status: Phase 0 Draft — will be refined after final stack & role grilling.

## 1. Product Summary
A hotel booking website for a **single hotel** with separate frontoffice and backoffice, manual bank-transfer payment, and a real-time availability calendar.

## 2. Users & Roles (4 Roles)
See details in `04-roles-permissions.md`. Summary:
- **Super Admin / Owner** — manages everything + views financial reports
- **Manager** — manages rooms, pricing, promotions, verifies bookings
- **Receptionist / Front Desk** — handles check-in/out, updates room status (clean/dirty/maintenance)
- **Customer** — registers, searches rooms, books, uploads proof, leaves reviews

## 3. Frontoffice Features (Customer)

### 3.1 Required (v1)
- [ ] Landing page: hero, room type listing, facilities, booking CTA
- [ ] Search availability: select `check_in` & `check_out` → see available rooms in real-time
- [ ] Room details: photos, description, facilities, price/night, capacity
- [ ] Booking flow: select room → enter guest details (name, phone, email) → create booking `pending_payment` → transfer instructions
- [ ] Upload transfer proof (image/PDF) → status `waiting_verification`
- [ ] Booking history + status tracking (pending, verified, checked_in, checked_out, cancelled, expired)
- [ ] Auth: register/login (email+password), forgot password, profile

### 3.2 Complete (v1 but can be incremental)
- [ ] Reviews & ratings per room type (only guests who have checked out can review)
- [ ] Voucher/promo codes (percentage discount, min. nights, expiry)
- [ ] Wishlist / favorites
- [ ] PDF invoice after verification
- [ ] Email notifications (booking created, verified, expired)

### 3.3 Down-to-Earth Example Scenario
> **Scenario A:** Ani wants a 2-night stay (Oct 10–12). She opens the website → selects dates → system checks: 2 Deluxe units still available → Ani books 1 Deluxe room → receives transfer instructions to BCA 123456 for IDR 1,000,000 → uploads proof → admin verifies within 10 minutes → status becomes `verified` → Ani receives an invoice email.

## 4. Backoffice Features (Management)

### 4.1 Rooms & Inventory
- [ ] CRUD for room types (name, description, capacity, price/night, number of units)
- [ ] CRUD for physical room units (e.g., Deluxe-101, Deluxe-102) + status: `available`, `occupied`, `dirty`, `maintenance`
- [ ] Occupancy calendar (view all bookings per date)
- [ ] Seasonal / weekend pricing (optional in v1)

### 4.2 Bookings & Operations
- [ ] Booking list: filter by status, date, room type
- [ ] Verify transfer proof (approve/reject + reason)
- [ ] Manual check-in / check-out + assign physical room unit
- [ ] Manual cancel & refund (record reason)
- [ ] Auto-expiry: bookings with `pending_payment` that remain unpaid for 2 hours → `expired`

### 4.3 Reports & Other
- [ ] Daily/monthly reports: occupancy rate, revenue, bookings per type
- [ ] Manage vouchers/promos
- [ ] Manage users & roles
- [ ] Audit log (who changed what, when)

## 5. Core Business Flow (Simplified)
```
Customer                    System                     Admin
   │                          │                          │
   ├─ search(dates) ─────────▶│                          │
   │◀─ list available rooms ───┤                          │
   ├─ create booking ────────▶│                          │
   │◀─ pending_payment ───────┤                          │
   ├─ upload proof ──────────▶│                          │
   │                          ├─ waiting_verification ──▶│
   │                          │◀─ verified/rejected ─────┤
   │◀─ notif verified ────────┤                          │
   │                          │                          ├─ check-in ──▶ occupied
   │                          │                          ├─ check-out ─▶ dirty → available
```

See `05-booking-flow-and-edge-cases.md` for detailed flow + edge cases.

## 6. Important Business Rules
| Rule | Example |
|------|---------|
| Check-in 14:00, check-out 12:00 | Booking Oct 10–12 = 2 nights, unit is free at 12:00 on the 12th |
| 1 booking = 1 room type, N nights | No mixing room types in a single booking (YAGNI) |
| Overlapping bookings prohibited | If a Deluxe room has 5 total units and all 5 are booked on date X, the 6th booking must be rejected |
| Price = price/night × nights − voucher discount | Voucher checks expiry & quota |
| Reviews only after check-out | Prevents fake reviews |

## 7. Non-Functional Requirements
- **Security:** IDOR/BOLA/BFLA blocked, SQLi via parameterized queries, XSS escaped, CSRF tokens, password hashing with bcrypt/argon2, JWT/cookie httpOnly.
- **Scalability:** Ready for multi-hotel later (add `hotel_id`), pagination, indexes on `check_in/out`.
- **Maintainability:** Clean Code, SRP, error handling + logging, testable.
- **ACID:** Booking creation must be a DB transaction (check availability + insert booking atomically).
- **Local-first:** Runs on M4 without Docker, ordered migrations.

## 8. Out of Scope v1
- Automatic payment gateway
- Multi-hotel
- Channel manager (Agoda)
- Mobile app

## 9. Definition of Done per Feature
- Has API + UI + validation + error handling + logging + at least 1 happy-path + 1 edge-case test.
- No happy-path-only — all edge cases in `05-...` must be handled.

> Next: `03-tech-stack-decision.md` to choose a stack without Next.js.
