# 13 — Phase 5: Invoice PDF Premium + Wishlist + CRUD + Cancel + Email Log + Audit (COMPLETE 2026-10-08 — YOLO)

> One infinite loop, all remaining gaps in one round, no git ops, premium WarmAura hierarchy, log-only email, wishlist heart, beautiful and complete.

## Summary

| Feature | Status | Evidence |
|---------|--------|----------|
| **Invoice PDF Premium WarmAura** | ✅ | `GET /api/bookings/:id/invoice` → 200 %PDF 3.3KB, header #8B5A2B cream gold line, hierarchy invoice# dates guest room nights price discount total, storage/invoices/invoice-35.pdf, verified/checked only else 409 |
| **Wishlist / Favorites** | ✅ | `POST /api/wishlist/toggle {room_type_id}` ↔ `{"added":true}`, `GET /api/wishlist` enriched, `DELETE`, heart filled #8B5A2B, /wishlist view grid |
| **Email Log Only** | ✅ | `backend/logs/email.log` written on create/verify/rejected/cancelled/expired via SendAsync goroutine mutex, slog + file append, non-blocking |
| **CRUD Room Types/Units** | ✅ | `POST/PUT/DELETE /api/admin/room-types` + `room-units`, soft delete deleted_at, validator $1, RBAC owner/manager, total_units denorm |
| **Cancel Booking** | ✅ | `PATCH /api/bookings/:id/cancel` owner of booking, only pending/waiting else 409 → cancelled, audit + email |
| **Audit Log View** | ✅ | `GET /api/admin/audit-logs` RBAC owner/manager, DataTable in AdminView |
| **Builds** | ✅ | `go vet 0`, `go test ok`, `npm build` 2053 modules 502ms Admin 342KB |
| **Live Full Loop** | ✅ | Wishlist add → booking → invoice 200 %PDF → email logged → cancel → Penthouse create |

## Invoice PDF Premium

**Lib:** `github.com/jung-kurt/gofpdf`

**Service `invoice.go`:** GenerateInvoice(bookingID) joins booking+user+room_type+voucher, nights = checkout-checkin, subtotal = price*nights, discount, total snapshot, proof_url. Renders:
- Header bg #8B5A2B rect + cream text XYZ HOTEL + Invoice #35 + gold line #C9A86A
- Hierarchy: Guest block (name email), Room block (type capacity), Dates (check_in/out nights), Price breakdown (Per night, Nights, Subtotal, Discount -10%, Total bold)
- Footer: Proof reference + Thank you + address, storage/invoices/invoice-{id}.pdf (3306 bytes %PDF)

**Handler:** `GET /api/bookings/:id/invoice` → Auth owner-or-admin (customer own only, owner/manager/receptionist all), validates status verified/checked_in/checked_out else 409 `Invoice is only available after verification`, serves PDF `Content-Type: application/pdf`, `Content-Disposition: inline; filename=invoice-35.pdf`, `Cache-Control: no-store`.

**Live:**
```bash
curl -o /tmp/invoice.pdf -w "%{http_code}" /api/bookings/35/invoice -H "Bearer $CUSTOMER"
# → 200 %PDF header, 3.3KB
curl /api/bookings/45/invoice (pending) → 409
```

## Wishlist

**Migration `003_wishlist.sql`:** wishlists (id, user_id FK, room_type_id FK, created_at, UNIQUE user+room_type), indexes, 004_room_soft.sql deleted_at on room_types.

**Repo/Handler:** Toggle idempotent (if exists → delete → added false, else insert → added true), ListEnriched joins room_types (avg_rating etc), Delete, $1, IDOR via user_id.

**Routes:**
- `POST /api/wishlist/toggle` {room_type_id} (Auth) → {added, message}
- `GET /api/wishlist` (Auth) → enriched array
- `DELETE /api/wishlist/:room_type_id` (Auth)

**Frontend:**
- HomeView heart button lucide Heart filled #8B5A2B if wished, toggle POST, toast Added/Removed, wishlist:updated event, badge count in App nav heart
- WishlistView.vue premium WarmAura grid, remove, toast, responsive
- MyBookings wishlist section
- App.vue nav Heart + badge

**Live:**
```bash
POST /api/wishlist/toggle {1} → {"added":true}
GET /api/wishlist → [{"room_type_id":1,"room_type":{Standard 5★}}]
```

## Email Log Only

**Service `email.go`:** Send(to, subject, body) → MkdirAll logs, mutex file append `backend/logs/email.log` (also logs/email.log fallback) with timestamp, slog. SendAsync via `go Send(...)` non-blocking.

**Templates warm and clear:**
- `Your XYZ Hotel booking #45 is received - pending payment` (booking created)
- `Your XYZ Hotel booking #35 has been verified` (verify)
- `Your XYZ Hotel booking #45 has been cancelled` (cancel)
- Expired via ticker: `Your XYZ Hotel booking #X has expired - no payment in 12 hours`

**Wiring:** availability CreateBooking after commit fetches email, verify handler after audit, cancel handler, expiry ticker fetches candidates before UPDATE then audit+email. Never SMTP, never blocks request.

**Live log:**
```
cat backend/logs/email.log
# [2026-10-08T19:17:12] to=customer@xyz-hotel.local subject=Your XYZ Hotel booking #45 is received...
# [2026-10-08T19:17:12] to=customer@xyz-hotel.local subject=Your XYZ Hotel booking #45 has been cancelled
```

## CRUD Rooms + Cancel + Audit

**Room Types CRUD:**
- `POST /api/admin/room-types` {name, description, capacity, price, total_units} → validator, $1, RBAC owner/manager, denorm total_units
- `PUT /api/admin/room-types/:id` → same
- `DELETE /api/admin/room-types/:id` → soft delete deleted_at, filter in List/Get
- `GET /api/admin/room-types` and public `GET /api/room-types` filtered

**Room Units CRUD:** similar, `POST /api/admin/room-units` {room_type_id, code, status}, `PUT`, `GET`

**Cancel:** `PATCH /api/bookings/:id/cancel` → Auth owner of booking, validates pending_payment/waiting_verification else 409, sets cancelled, audit, email.

**Audit Log:** repo/audit.go List ordered desc, handler GET /api/admin/audit-logs?limit&offset RBAC owner/manager, AdminView Audit tab DataTable.

**Live:**
```bash
POST /api/admin/room-types {"name":"Penthouse",...} (owner) → 201 {"id":10}
PATCH /api/bookings/45/cancel (customer own pending) → {"status":"cancelled"}
GET /api/admin/audit-logs (owner) → audit entries
```

## Frontend Hierarchy WarmAura Premium

**Wishlist heart:** Heart lucide filled #8B5A2B when wished, outline #6B7280 otherwise, hover scale 105%, gold #C9A86A on hover, count badge cream bg gold text.

**Invoice button:** Download Invoice (lucide Download) #8B5A2B when verified/checked, blob download, toast.

**Admin CRUD:** Tabs Bookings/Room Types/Room Units/Audit, Room Types Dialog (InputText, InputNumber) WarmAura #8B5A2B primary, gold dividers, rounded-2xl, DataTable header #1A3A4A.

**Hierarchy:** Playfair Display for invoice # and room title, Inter for body, spacing xl, contrast 4.5:1, grid 1/2/4.

## Verification

```bash
go vet ./...          # 0
go test ./...         # ok handler+service
npm run build         # 2053 modules 502ms
curl wishlist toggle → added true → GET wishlist enriched
curl invoice 35 → 200 %PDF
cat backend/logs/email.log → 2 entries logged
curl cancel pending → cancelled
curl room-types create Penthouse → 201
```

## DoD Phase 5

- [x] Invoice PDF premium WarmAura hierarchy
- [x] Wishlist heart toggle + view
- [x] Email log only non-blocking
- [x] CRUD room types/units soft delete + cancel + audit log
- [x] Frontend beautiful hierarchy + build OK
- [x] No git ops (YOLO)

## How to Run Phase 5

```bash
cd backend && go run ./cmd/server  # :8080
# Wishlist
curl -X POST :8080/api/wishlist/toggle -H "Bearer $CUSTOMER" -d '{"room_type_id":1}'
# Invoice
curl :8080/api/bookings/35/invoice -H "Bearer $CUSTOMER" -o invoice.pdf
# Email log
cat backend/logs/email.log
# Cancel
curl -X PATCH :8080/api/bookings/45/cancel -H "Bearer $CUSTOMER"
# CRUD
curl -X POST :8080/api/admin/room-types -H "Bearer $OWNER" -d '{"name":"Penthouse",...}'
cd frontend && npm run dev  # :5173 heart, wishlist, invoice, admin CRUD
```

> All SRP, $1 safe, WarmAura #8B5A2B, log-only, heart toggle, premium PDF.
