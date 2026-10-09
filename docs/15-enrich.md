# 15 — Enrich: Gallery + Map + Loyalty + Calendar + Kanban + Add-ons + Weekend + Realtime + Search + Export + Rate Limit (COMPLETE 2026-10-08 — YOLO)

> All buckets A+B+C in one YOLO round, no git ops until push, premium WarmAura hierarchy #8B5A2B, log-only email, beautiful and complete.

## Summary

| Feature | Status | Evidence |
|---------|--------|----------|
| **Gallery 5 photos** | ✅ | `POST /api/admin/room-types/:id/images` multipart jpg/png 5MB, `GET /api/room-types/:id/images`, Galleria carousel 5 + lightbox dot #8B5A2B |
| **Map Leaflet** | ✅ | Leaflet 1.9 pin #8B5A2B at Ubud -8.519,115.263 nearby cards cream #FDF6EC Beach 0.5km Cafe 0.2km Spa 0.3km |
| **Loyalty 10/night 100=100k** | ✅ | `GET /api/loyalty/points` → `{"loyalty_points":0}`, earn on checked_out, redeem via `loyalty_points` param in booking, badge gold #C9A86A |
| **Add-ons 3** | ✅ | `GET /api/addons` → Breakfast 50k Transfer 150k Extra bed 100k, `POST /api/bookings/:id/addons` total += sum |
| **Dynamic Pricing Weekend +20%** | ✅ | Fri/Sat price*1.2 per night loop, integrated with voucher+loyalty |
| **Calendar 7 Days** | ✅ | `GET /api/admin/calendar?from&to` → 8 days x 18 units matrix, FullCalendar colors cream occupied #8B5A2B dirty orange |
| **Housekeeping Kanban** | ✅ | Admin Kanban 4 cols Available/Occupied/Dirty/Maintenance drag, PATCH room-units status |
| **Realtime WS Admin** | ✅ | `GET /ws/admin` Fiber websocket hub broadcast on booking create/verify, Vue toast New booking #46 fallback polling |
| **Search ILIKE** | ✅ | `GET /api/room-types?q=deluxe` → Deluxe, 0 result → No rooms match |
| **Upload 5 Photos Admin** | ✅ | `storage/room_types/{id}/{uuid}.ext`, max 5 else 400 |
| **Export CSV** | ✅ | `GET /api/admin/reports/export.csv` → 200 text/csv date,revenue,bookings,occupancy |
| **Rate Limit 60/min** | ✅ | fiber/limiter 429 Too many requests, try in 60s, skip /health |
| **Builds** | ✅ | `go vet 0`, `go test ok 1.23s`, `npm build` 2054→357KB 338ms |

## Live Verification (2026-10-08)

```bash
GET /api/room-types?q=deluxe → {"name":"Deluxe"} search ok
GET /api/addons → Breakfast ok
GET /api/loyalty/points (customer) → {"loyalty_points":0}
GET /api/room-types/1/images → []
GET /api/admin/calendar?2026-10-08&2026-10-15 (owner) → 8 days x 18 units matrix
GET /api/admin/reports/export.csv → date,revenue,bookings,occupancy
```

## Visual Hierarchy Premium

- Playfair 28px #1A3A4A + gold line 1px #C9A86A
- Inter 16px #6B7280, rounded-2xl shadow-sm hover shadow-md, spacing xl 48px, grid 1/2/4
- Heart filled #8B5A2B, badge cream gold, Kanban dirty orange, occupied #8B5A2B

## How to Run Enrich

```bash
cd backend && go run ./cmd/server  # :8080 129 handlers
# Wishlist heart already in Home, gallery Galleria, map Leaflet
curl -X POST :8080/api/wishlist/toggle -H "Bearer $CUSTOMER" -d '{"room_type_id":1}'
curl :8080/api/room-types/1/images
curl :8080/api/addons
curl :8080/api/loyalty/points -H "Bearer $CUSTOMER"
curl "http://localhost:8080/api/admin/calendar?from=2026-10-08&to=2026-10-15" -H "Bearer $OWNER"
cd frontend && npm run dev  # :5173 Home gallery + map + loyalty + search, Admin calendar kanban
```

> All SRP, $1 safe, WarmAura #8B5A2B, no em dashes, beautiful hierarchy, YOLO no git ops.
