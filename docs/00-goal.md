# 00 — Goal & Vision xyz-hotel

## What is xyz-hotel?
A **website-based** hotel booking application with 2 sides:

```
┌─────────────┐        ┌──────────────┐        ┌─────────────┐
│ Frontoffice │───────▶│  Backend API │◀───────│ Backoffice  │
│ (Customer)  │        │              │        │ (Management)│
└─────────────┘        └──────────────┘        └─────────────┘
      │                                              │
  View rooms,                            Manage rooms, prices,
  check dates,                           verify bookings,
  book, upload                           reports, vouchers,
  payment proof                          change room status
```

**Down-to-earth analogy:**
- Frontoffice = like opening Traveloka → searching for a hotel → picking dates → booking
- Backoffice = like the receptionist dashboard in the hotel lobby → seeing who checks in today, which rooms are dirty/clean

## Primary Goals
1. Customers can book rooms **without calling the receptionist**.
2. Management can manage rooms, prices, bookings, and reports **without manual Excel**.
3. Rafie's personal portfolio — creative freedom, high quality, no deadline, as a skill showcase.

## Phase 0 Constraints (agreed 2026-10-08)
| Aspect | Decision | Reason |
|-------|----------|--------|
| Number of hotels | **One hotel for now** | YAGNI — don't over-engineer multi-hotel early, but keep DB ready to scale |
| Feature level | **Complete** (search, availability, booking, review, voucher, reports) | Portfolio must be impressive, but implemented incrementally |
| Roles | **4 roles** | Clear separation of responsibilities required |
| Payment | **Manual transfer + proof upload** | Simple, no gateway fees, suitable for a local portfolio |
| Availability | **MUST be real-time** | Core hotel logic — if wrong, double bookings happen |
| Tech stack | **DO NOT use Next.js** | User's hard preference, respect it |
| Environment | **Local only on MB Air M4** | Postgres 18.4 / MySQL 9.6, PHP 8.2, Node 26, no Docker |
| Repo | Local `main` branch, push to GitHub later | No worktrees, work directly on local branch |

## What is NOT a Goal (Non-Goals) in v1
- ❌ Automatic payment gateway (Midtrans/Xendit) — v2
- ❌ Multi-hotel / multi-branch — v2
- ❌ Native mobile app — website first
- ❌ Channel manager (sync to Agoda/Booking.com) — out of scope

## Success Criteria (When is it considered successful?)
- Customers can complete a booking from searching dates to uploading proof without admin assistance.
- Admin can verify a booking in < 5 minutes and change room status.
- Double bookings on the same date never occur (ACID is maintained).
- Daily/monthly reports are generated accurately.

## Definition of Done — Phase 0
- [x] `xyz-hotel` folder + git init
- [x] These 6 docs files completed
- [ ] Final tech stack without Next.js agreed
- [ ] 4 roles + booking flow agreed
- [ ] Ready for Phase 1 scaffolding

> Next: read `01-brainstorming-log.md` for the full conversation context.
