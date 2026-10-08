# 06 — Context & Constraints (Portfolio, Local, Security)

> This file answers Q5: "You decide, be as creative as you want" — but still measured.

## Personal Portfolio Context
- **No deadline, no client, no boss** — you stand on your own two feet. That means: quality > speed. We are not chasing sprints; we are chasing a showcase you can be proud of on your CV/GitHub.
- **Free to be creative:** Room types, pricing, amenities, design — the agent may decide freely as long as it remains realistic for a 3-star hotel business.
- **Room types — LOCKED 2026-10-08 (to be seeded via `database/seeders/RoomSeeder.php`):**

| Type | Capacity | Bed | Amenities | Price/night | Units |
|------|----------|-----|-----------|-------------|-------|
| Standard | 2 | 1 Queen | AC, TV, En-suite bathroom | Rp 350,000 | 8 |
| Deluxe | 2 | 1 Queen + Sofa | + Balcony, Mini fridge | Rp 550,000 | 5 |
| Family | 4 | 2 Queen | + Mini kitchen, 2 bedrooms | Rp 850,000 | 3 |
| Suite | 2 | 1 King | + Living room, Jacuzzi | Rp 1,250,000 | 2 |

> Total 18 physical units (STD-101..108, DLX-201..205, FAM-301..303, STE-401..402). Price is snapshotted to `bookings.total_price` at booking time.

## Principles to Uphold (No Future Debt)
| Principle | Application in xyz-hotel |
|-----------|--------------------------|
| **YAGNI** | Do not build multi-hotel, multi-room per booking, or a payment gateway in v1. |
| **DRY** | Availability logic lives in one service `AvailabilityService`, do not copy-paste across controllers. |
| **KISS** | Keep date validation simple: `check_out > check_in`, do not build a complex pricing engine yet. |
| **SRP** | `BookingController` handles HTTP only, `BookingService` handles transactions, `BookingPolicy` handles authorization. |
| **SOLID** | Use dependency injection for services, not `new` inside controllers. |
| **ACID** | Booking creation within a single transaction. |
| **Clean Code** | Variable names like `occupiedUnits` not `x`, functions < 30 lines. |
| **Testable** | Availability logic can be unit-tested without HTTP. |

## Security — No Leaks Allowed (Checklist)

| Threat | Prevention in xyz-hotel |
|--------|--------------------------|
| **SQL Injection** | Eloquent ORM + parameterized queries, never raw string concatenation |
| **IDOR / BOLA** | Policy check `booking.user_id == auth.id` |
| **BFLA** | Role middleware on every route group |
| **Broken Auth** | Argon2/bcrypt hashing, httpOnly cookies, rate limit 5x/minute on login |
| **XSS** | Escape output in Vue (`{{ }}` auto-escaped), input validation |
| **CSRF** | Laravel CSRF token automatically for Inertia |
| **Sensitive Data** | Do not log passwords, do not expose `user.email` to other customers |
| **Security Misconfig** | Do not commit `.env`, `APP_DEBUG=false` in prod, generic error messages |
| **SSRF** | Validate upload URLs (local storage only, do not fetch external URLs) |

## Scalability & Maintainability (Portfolio Must Be Long-Lived)
- **Pagination** on booking/room listings (do not `SELECT *` without a limit).
- **DB indexes:** `bookings(check_in, check_out, room_type_id, status)`, `users(email)`.
- **Soft deletes** for rooms & users (do not hard-delete — booking history is needed).
- **Audit log** in a separate table.
- **API documentation** in `docs/02-api.md` (to be created in Phase 1).
- **Structured logging:** `Log::info('booking.created', ['id'=>..])`.

## Creativity Allowed (Because No Rules)
- Free landing page design — can use a warm hotel theme (gold, cream) distinct from the blue of xyz-haircut.
- Bonus portfolio features: hotel location map, photo gallery, FAQ, WhatsApp contact.
- Reports can use charts (Chart.js) — a chance to showcase skills.

## Global Definition of Done (Portfolio Grade)
- [ ] Clean code, README with local setup instructions
- [ ] No remaining TODO/FIXME
- [ ] Screenshots / demo GIF in README
- [ ] Public GitHub repo, clean commit history
- [ ] No committed secrets

## Next Steps After Phase 0
1. Lock tech stack (answer grill `03`)
2. Lock roles & flow (answer grill `04` & `05`)
3. `git remote add` + push to GitHub (on your request)
4. Phase 1: Laravel scaffolding + DB migrations

> If you agree with the room types above, say "agree to room types". If you want changes, tell us what you'd like.
