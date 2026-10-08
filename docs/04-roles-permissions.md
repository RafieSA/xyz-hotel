# 04 — Roles & Permissions (4 Roles)

> Grilling Q3 — 4 roles for a single hotel. Made down-to-earth with examples.

## Role List

| # | Role | Who in Real Life | Analogy |
|---|------|------------------|---------|
| 1 | **Super Admin / Owner** | Hotel owner | The big boss, sees the money, manages everything |
| 2 | **Manager** | Operations manager | Manages rooms, pricing, promos, verifies bookings |
| 3 | **Receptionist** | Front desk / receptionist | Check-in/out, updates room status to clean/dirty |
| 4 | **Customer** | Guest / customer | Books, pays, reviews |

> Role names can be changed (e.g., Owner → Super Admin). What matters is that permissions are clear.

## Permission Matrix

| Feature | Owner | Manager | Receptionist | Customer |
|---------|-------|---------|:------------:|:--------:|
| **Manage users & roles** | ✅ | ❌ | ❌ | ❌ |
| **CRUD room types & physical units** | ✅ | ✅ | ❌ | ❌ |
| **Change pricing & promos/vouchers** | ✅ | ✅ | ❌ | ❌ |
| **View all bookings** | ✅ | ✅ | ✅ | ❌ (own only) |
| **Verify transfer proof** | ✅ | ✅ | ❌ | ❌ |
| **Check-in / Check-out** | ✅ | ✅ | ✅ | ❌ |
| **Change room status (dirty/maintenance)** | ✅ | ✅ | ✅ | ❌ |
| **View revenue & occupancy reports** | ✅ | ✅ | ❌ | ❌ |
| **Create booking & upload proof** | ❌ | ❌ | ❌ | ✅ |
| **Leave a review** | ❌ | ❌ | ❌ | ✅ (only after check-out) |
| **Manage own profile** | ✅ | ✅ | ✅ | ✅ |

## Security Rules (Anti IDOR/BOLA/BFLA)

| Threat | Example | Prevention |
|--------|---------|------------|
| **IDOR** | Customer id=5 tries to view booking id=99 belonging to someone else via `/bookings/99` | Check `booking.user_id == auth.id` OR admin role — in Policy |
| **BOLA** | Customer tries `POST /rooms` (create room) | Middleware `role:manager,owner` — reject with 403 |
| **BFLA** | Receptionist tries to open `/reports/revenue` | Middleware `role:owner,manager` — reject with 403 |
| **Broken Auth** | Token is stolen | Password hashed with argon2/bcrypt, httpOnly cookie / JWT 15 min + refresh, login rate limiting |

**Example code (Laravel Policy) — illustration:**
```php
// BookingPolicy.php
public function view(User $user, Booking $booking): bool {
    return $user->id === $booking->user_id || $user->hasRole(['owner','manager','receptionist']);
}
public function verify(User $user): bool {
    return $user->hasRole(['owner','manager']);
}
```

## Role Edge Cases
| Case | Handling |
|------|----------|
| Manager resigns, account still active | Owner deactivates the user; all bookings remain intact (soft-delete user, not hard delete) |
| Receptionist tries to verify a booking (not authorized) | API returns 403 + logs `unauthorized attempt` |
| Customer tries to self check-in | Rejected — only receptionist/manager/owner |
| Owner also wants to be a customer | Allowed — 1 user has 1 primary role, but can create a separate customer account (don't mix roles) |

## Best Practices
- **1 user = 1 role** in v1 (YAGNI — don't do many-to-many yet).
- Store `role` in `users.role` enum: `owner, manager, receptionist, customer`.
- Every endpoint checks **authentication first, then authorization** (not the other way around).
- Log every sensitive action: `who, what, when, before, after` (audit log).

## Grilling Decisions — LOCKED 2026-10-08
1. **Agree with the role names above?** ✅ **Agreed** (Rafie 2026-10-08)
2. **Can Manager view financial reports?** ✅ **Yes, Manager + Owner**
3. **Must customer log in before searching?** ✅ **No login required for search**, login only required to book
