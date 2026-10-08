# 05 — Booking Flow & Edge Cases (The Heart of the Hotel)

> Grill Q6 — Availability is mandatory. This is the most critical section and must be ACID.

## Normal Flow (Happy Path) — Visual

```
Date: Oct 10-12 (2 nights), Type: Deluxe, Total units: 5

Customer                Frontend                Backend (DB Transaction)         Admin
   │                       │                           │                          │
   ├─ select dates 10-12 ─▶│                           │                          │
   │                       ├─ GET /availability ──────▶│                          │
   │                       │   ┌─ calculate:           │                          │
   │                       │   │  SELECT COUNT(*)      │                          │
   │                       │   │  FROM bookings        │                          │
   │                       │   │  WHERE room_type=Deluxe│                          │
   │                       │   │  AND status IN (verified, checked_in)            │
   │                       │   │  AND overlap(dates)   │                          │
   │                       │◀──┤  → 3 occupied, 2 available │                     │
   │◀─ show 2 available ───┤                           │                          │
   ├─ book ───────────────▶├─ POST /bookings ─────────▶│                          │
   │                       │   ┌─ BEGIN TRANSACTION     │                          │
   │                       │   │  1. check availability again (FOR UPDATE)        │
   │                       │   │  2. insert booking pending_payment               │
   │                       │   │  3. COMMIT / ROLLBACK  │                          │
   │                       │◀──┤  → pending_payment     │                          │
   │◀─ transfer instructions ┤                           │                          │
   ├─ upload proof ───────▶├─ POST /bookings/1/proof ─▶│                          │
   │                       │                           ├─ waiting_verification ──▶│
   │                       │                           │◀─ approve ───────────────┤
   │◀─ verified + invoice ─┤                           │                          │
   │                       │                           │                          ├─ check-in → occupied
   │                       │                           │                          ├─ check-out → dirty → available
```

**Booking Status:**
```
pending_payment → waiting_verification → verified → checked_in → checked_out
       │                    │               │
       ├─ expired (12 hours)└─ rejected ────┘
       └─ cancelled (by customer before verification)
```

## Overlap Rules (Key to Preventing Double Bookings)

**Overlap definition:** Booking A `[10,12)` and Booking B `[11,13)` → overlap on the 11th.

**DB Formula (Postgres/MySQL):**
```sql
-- Check whether there is an overlap with an already verified/checked_in booking
WHERE room_type_id = :type
  AND status IN ('verified','checked_in')
  AND check_in < :new_check_out
  AND check_out > :new_check_in
```

**Calculation example:**
- Deluxe total units = 5
- There are already 5 verified bookings overlapping on the 11th → **reject** new booking (403, "Rooms fully booked")
- Only 4 overlap → **allowed** (1 slot remaining)

## Extreme Edge Cases (Must Be Handled — Happy-Path-Only Is Forbidden)

| # | Edge Case | Example | Handling |
|---|-----------|---------|----------|
| 1 | **Race condition** — 2 people booking at the exact same second | At 10:00:00, Ani & Budi both book Deluxe for the same dates, 1 slot left | Use `DB::transaction()` + `SELECT ... FOR UPDATE` / unique constraint. One succeeds, one fails with a clear message. |
| 2 | **Booking expired** | Ani books but does not upload proof within 12 hours | Cron/job every minute: `pending_payment` & `created_at < now-12h` → `expired`. Slot is released. |
| 3 | **Fake / incorrect proof** | Uploads empty photo or insufficient amount | Admin rejects with reason, status `rejected`, customer may re-upload (max 3x) or booking becomes `cancelled`. |
| 4 | **Check-in without verification** | Receptionist attempts to check in a pending booking | Rejected — only `verified` can become `checked_in`. Validate on the backend, not just the frontend. |
| 5 | **Late check-out** | Guest checks out at 14:00 (past 12:00) | Status remains `checked_in`, penalty calculated manually in the back office (recorded in report). |
| 6 | **Cancellation after verification** | Guest is verified but wants to cancel the day before arrival | Manager cancels → status `cancelled`, unit released, refund handled manually (recorded). |
| 7 | **Voucher expired / quota exhausted** | Code PROMO10 used 101 times though quota is 100 | Check voucher `expiry` & `used_count < quota` within the same transaction. |
| 8 | **Invalid dates** | check_in = check_out, or check_in in the past | Validation: `check_out > check_in`, `check_in >= today`, `nights >=1`. |
| 9 | **Over capacity** | Deluxe capacity is 2 guests, guest enters 5 | Validation: `guests <= capacity` + `guests >=1`. |
| 10 | **Malicious file upload** | Uploads `.php` or 50 MB file | Validate MIME `jpg,png,pdf`, max 5 MB, rename file, store in `storage/app/private` (not public), scan. |
| 11 | **IDOR — viewing another user's booking** | Customer A tries GET /bookings/99 owned by B | Policy: `booking.user_id == auth.id` or admin role. |
| 12 | **Price changed mid-booking** | Deluxe price rises from 500k to 600k after Ani books | Price is **snapshotted** to `bookings.total_price` at creation — does not follow the new price. |

## ACID Best Practices & Logging

```php
// Laravel illustration — MUST use a transaction
DB::transaction(function () use ($data) {
    // 1. Lock relevant rows
    $occupied = Booking::where('room_type_id', $data->type)
        ->whereIn('status', ['verified','checked_in'])
        ->where('check_in', '<', $data->check_out)
        ->where('check_out', '>', $data->check_in)
        ->lockForUpdate()->count();

    $totalUnits = RoomType::find($data->type)->total_units;
    if ($occupied >= $totalUnits) {
        throw new \Exception('Rooms fully booked for those dates');
    }

    // 2. Insert booking
    $booking = Booking::create([... , 'total_price' => $priceSnapshot]);

    // 3. Audit log
    AuditLog::create(['user_id' => auth()->id(), 'action' => 'booking_created']);

    return $booking;
});
```

- **Logging:** Log every status change: `booking_id, from, to, by, at, reason`.
- **Error handling:** Do not use empty `try-catch` — log to `storage/logs/laravel.log` + return a user-friendly message.

## Grill Decisions — LOCKED 2026-10-08
1. **How long until expiry?** ✅ **12 hours** (Rafie's decision — `pending_payment` & `created_at < now-12h` → `expired`)
2. **What are check-in/out times?** ✅ **14:00 / 12:00** (approved)
3. **How many rooms per booking?** ✅ **1 type, 1 unit, N nights** (approved, YAGNI)
> Next: `06-context-and-constraints.md` for portfolio & local-first context.
