# 11 — Phase 3: Voucher, Proof Upload, Verification & Check-in/out (COMPLETE 2026-10-08)

> One full round: voucher discount + multipart proof upload (MIME + 5MB) + admin verification + check-in/out with room-unit assignment (FOR UPDATE). Live curl verified end-to-end.

## Summary

| Item | Status | Evidence |
|------|--------|----------|
| **Voucher CRUD** | ✅ | POST /api/admin/vouchers (owner/manager) + GET validate, discount applied transactionally |
| **Proof Upload** | ✅ | POST /api/bookings/:id/proof multipart, MIME magic (JPEG/PNG/PDF), 5MB, IDOR, status pending→waiting |
| **Verification** | ✅ | PATCH /api/admin/bookings/:id/verify (waiting→verified/rejected), RBAC owner/manager |
| **Check-in/out** | ✅ | PATCH checkin (verified→checked_in, assigns unit FOR UPDATE, occupied), checkout (checked_in→checked_out, dirty), PATCH room-units status |
| **Frontend Wiring** | ✅ | HomeView voucher input, MyBookings upload, AdminView verify/checkin/checkout, room-units table |
| **Builds** | ✅ | `go vet 0`, `go test ok` (service+handler), `npm run build` 2044 modules OK |
| **Live Flow** | ✅ | create with voucher 700k→630k (10% off) → upload PNG → verify → checkin unit 1 → checkout dirty |

## Voucher

**Model:** `vouchers` (code unique, discount_percent 0-100, min_nights, quota nullable, used_count, expires_at)

**Repo `voucher.go`:** FindByCode, FindByCodeForUpdate (FOR UPDATE), Create, IncrementUsedCountTx, List — all $1 safe.

**Service `voucher.go`:** ValidateAndApply(code, roomTypeID, checkIn, checkOut) checks expiry, quota (used_count < quota), min_nights (nights >= min_nights), returns discount + voucherID; CalculateDiscountedPrice(price, nights, discount) = price*nights*(100-discount)/100.

**Handlers `voucher.go`:**
- `POST /api/admin/vouchers` → RBAC owner/manager, validates discount 0-100, code unique → 201
- `GET /api/admin/vouchers` → list
- `GET /api/vouchers/validate?code&room_type_id&check_in&check_out` → public preview → {discount_percent, discounted_total, original_total, nights} or 400/404

**Booking Integration:** `AvailabilityService.CreateBooking` now accepts voucherCode, validates in same DB transaction (FOR UPDATE voucher row), increments used_count atomically, snapshots discounted total_price.

**Tests `voucher_test.go`:** expiry, quota, min_nights, CalculateDiscountedPrice, DiscountedTotal — table-driven, all pass.

**Live:**
```bash
POST /api/admin/vouchers {"code":"WELCOME10","discount_percent":10,"min_nights":2,"quota":100} → 201
GET /api/vouchers/validate?code=WELCOME10&room_type_id=1&check_in=2026-10-20&check_out=2026-10-22
→ {"discount_percent":10,"discounted_total":630000,"original_total":700000,"nights":2}
POST /api/bookings {"room_type_id":1,"check_in":"2026-10-25","check_out":"2026-10-27","voucher_code":"WELCOME10"}
→ {"total_price":630000, "voucher_id":1} # 700k *0.9
```

## Proof Upload & Verification

**Storage:** `backend/storage/uploads/bookings/` (gitignored `uploads/*`, .gitkeep), created at startup, Fiber BodyLimit 6MB.

**Repo `booking.go`:** GetByIDTx, UpdateProofURLTx, UpdateStatusTx — FOR UPDATE.

**Handler `booking.go` UploadProof:**
- `POST /api/bookings/:id/proof` → Auth + IDOR (booking.user_id == auth.id, else 403)
- Validates `c.FormFile("proof")`: size ≤5MB before reading, ext ∈ {jpg,jpeg,png,pdf} case-insensitive, MIME via magic bytes (JPEG FF D8 FF, PNG 89 50 4E 47, PDF %PDF)
- Renames to `bookings/{id}_{uuid}.ext`, saves to `storage/uploads`, updates proof_url + status pending_payment→waiting_verification (only if pending), audit log, slog.

**Handler VerifyBooking:**
- `PATCH /api/admin/bookings/:id/verify` → Auth + RequireRole(owner,manager)
- Body `{action: "verified"|"rejected", reject_reason?: string}` — reject_reason required if rejected
- Validates status == waiting_verification else 409, updates to verified/rejected, audit, slog.

**Tests `upload_test.go`:** jpg/jpeg/png/pdf pass, php/exe fail, magic mismatch fail, 6MB fail, 5MB pass, empty/txt fail.

**Live:**
```bash
curl -X POST :8083/api/bookings/14/proof -H "Bearer $CUSTOMER" -F "proof=@/tmp/test.png"
→ {"status":"waiting_verification","proof_url":"bookings/14_6af05eed...png"} # MIME PNG ok

curl -X PATCH :8083/api/admin/bookings/14/verify -H "Bearer $OWNER" -d '{"action":"verified"}'
→ {"status":"verified"}
```

## Check-in/out & Room Units

**Repo `room.go`:** FindAvailableUnitForTypeTx (SELECT id FOR UPDATE LIMIT 1 where status available), GetUnitByIDTx, UpdateStatusTx, ListAllUnits.

**Service `booking_ops.go`:** BookingOpsService with CheckIn (verified→checked_in: picks available unit FOR UPDATE, sets bookings.room_unit_id + checked_in, sets room_units occupied), CheckOut (checked_in→checked_out: sets dirty), UpdateRoomUnitStatus (validates transition via IsValidRoomStatusTransition, FOR UPDATE).

**Handlers `booking.go`:**
- `PATCH /api/admin/bookings/:id/checkin` → Auth + RequireRole(owner,manager,receptionist), 409 if not verified, 409 if no available unit
- `PATCH /api/admin/bookings/:id/checkout` → same RBAC, 409 if not checked_in
- `PATCH /api/admin/room-units/:id/status` → body {status: available|occupied|dirty|maintenance}, validates via IsValidRoomStatusTransition, FOR UPDATE
- `GET /api/admin/room-units` → list

**Valid Transitions:** available→occupied/dirty/maintenance, occupied→dirty, dirty→available/maintenance, maintenance→available.

**Live:**
```bash
PATCH :8083/api/admin/bookings/14/checkin -H "Bearer $OWNER"
→ {"status":"checked_in","room_unit_id":1} # unit STD-101 occupied

PATCH :8083/api/admin/bookings/14/checkout -H "Bearer $OWNER"
→ {"status":"checked_out"} # unit 1 → dirty

PATCH :8083/api/admin/room-units/1/status -d '{"status":"available"}' -H "Bearer $OWNER"
→ {"status":"available"} # dirty → available

GET :8083/api/admin/room-units -H "Bearer $OWNER" → 18 units
```

**Security:**
- IDOR: proof upload checks owner
- BOLA: voucher create admin only
- BFLA: verify owner/manager, checkin/out owner/manager/receptionist, room status same
- $1 safe, validator oneof, slog audit

## Frontend Wiring

**Files:**
- `frontend/src/api/client.js` → axios + Bearer
- `frontend/src/views/HomeView.vue` → voucher InputText + Validate button → GET /api/vouchers/validate, shows discount % + discounted total (line-through original), nightsCount, passes voucher_code on CreateBooking
- `frontend/src/views/MyBookingsView.vue` → GET /api/bookings, DataTable with Tag status, proof image preview (storage/uploads), upload form multipart (file input jpg/png/pdf max 5MB → POST /api/bookings/:id/proof)
- `frontend/src/views/AdminView.vue` → DataTable bookings with Tag, Verify/Reject buttons (if waiting), Check-In (if verified), Check-Out (if checked_in), room_units DataTable with status Select→ PATCH, toast, Dialog proof image
- `frontend/src/router/index.js` → /bookings → MyBookingsView
- `frontend/src/App.vue` → nav Booking Saya

**Build:** `npm run build` 2044 modules, ✓ built 253ms (HomeView 115KB, AdminView 413KB, MyBookings 7.59KB).

## Verification (Factual 2026-10-08)

```bash
cd backend && go vet ./...          # 0
cd backend && go test ./...         # ok service+handler (voucher, upload, transition)
cd frontend && npm run build        # 2044 modules OK
# Live full flow:
POST /api/admin/vouchers WELCOME10 → 201
GET /api/vouchers/validate → 630k/700k
POST /api/bookings voucher_code WELCOME10 → 630k
POST /api/bookings/:id/proof PNG 1x1 → waiting_verification bookings/14_*.png
PATCH /api/admin/bookings/:id/verify verified → verified
PATCH /api/admin/bookings/:id/checkin → checked_in unit 1 occupied
PATCH /api/admin/bookings/:id/checkout → checked_out unit 1 dirty
PATCH /api/admin/room-units/1/status available → available
```

## DoD Phase 3

- [x] Voucher CRUD + validation transactional + discount snapshot
- [x] Proof upload MIME + 5MB + IDOR + status pending→waiting
- [x] Verification waiting→verified/rejected (RBAC owner/manager)
- [x] Check-in/out with unit assignment FOR UPDATE + room status transitions
- [x] Frontend voucher input, upload, admin actions + build OK
- [ ] Next: reviews + reports + admin dashboard polish (Phase 4)

## How to Run Phase 3

```bash
cd backend && go run ./cmd/server  # :8080
# Voucher (owner)
curl -X POST :8080/api/admin/vouchers -H "Bearer $OWNER" -d '{"code":"WELCOME10","discount_percent":10,"min_nights":2,"quota":100}'
# Booking with voucher
curl -X POST :8080/api/bookings -H "Bearer $CUSTOMER" -d '{"room_type_id":1,"check_in":"2026-10-25","check_out":"2026-10-27","voucher_code":"WELCOME10"}'
# Upload proof
curl -X POST :8080/api/bookings/<id>/proof -H "Bearer $CUSTOMER" -F "proof=@/tmp/test.png"
# Verify
curl -X PATCH :8080/api/admin/bookings/<id>/verify -H "Bearer $OWNER" -d '{"action":"verified"}'
# Check-in/out
curl -X PATCH :8080/api/admin/bookings/<id>/checkin -H "Bearer $OWNER"
curl -X PATCH :8080/api/admin/bookings/<id>/checkout -H "Bearer $OWNER"
```

> All SRP, $1 safe, MIME magic, 5MB, RBAC, FOR UPDATE, WarmAura #8B5A2B.
