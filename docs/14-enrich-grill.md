# 14 — Enrich Grill: Guest Wow + Ops Excellence + Tech Proper (LOCKED 2026-10-08)

> Grill for all remaining gaps, priority enrich focus. Buckets A+B+C, premium WarmAura hierarchy #8B5A2B, log-only email, wishlist required. Document then YOLO all in one round, no git ops until push.

## Decision Locked

**Scope:** Enrich all three buckets in one YOLO round, no git ops until push.

| Bucket | Features Locked | Why |
|--------|-----------------|-----|
| **A Guest Wow** | Gallery premium (5 photos per room type, carousel + lightbox), Map Leaflet + nearby, Loyalty points (10 points per night, 100 points = IDR 100k discount) | Guest sees hotel as premium and alive, not 4 cards, loyalty showcases business logic |
| **B Ops Excellence** | Occupancy calendar drag (7 days x 18 units, colors), Housekeeping Kanban (Available/Occupied/Dirty/Maintenance drag), Add-ons (Breakfast 50k, Airport transfer 150k, Extra bed 100k), Dynamic pricing weekend +20% | Admin never uses Excel again, ops wow, showcase drag and pricing |
| **C Tech Proper** | Realtime WebSocket admin (new booking toast), Full-text search ILIKE, Room photo upload (5 per type, admin), Export reports CSV, Rate limit 60/min | Proper engineering, portfolio engineering depth, no SMTP |

## Bucket A Grill

### Gallery Premium

| Aspect | Choice | Tradeoff | Edge Case | Rationale |
|--------|--------|----------|-----------|-----------|
| **Library** | PrimeVue Galleria + vue-easy-lightbox | Galleria is PrimeVue native, lightbox 300ms fade, 2KB | Photo 404 → placeholder cream #FDF6EC + lucide Image, lazy load | Galleria fits Aura, no extra React lib |
| **Photos** | 5 per room type, seeded Unsplash, admin can replace via upload | 5 is premium but not heavy, 20 total images, lazy + width 600 | 6th photo → 400 Max 5 photos | 5 balances beauty and load |
| **Placement** | Below Choose your room, section Our Spaces (Playfair 28px #1A3A4A + gold line 1px) → full-width rounded-2xl carousel | Full-width is hero for rooms, not grid | Mobile carousel swipe, desktop dots | Hierarchy warm resort, not generic grid |

**Hierarchy:** Title Playfair 28px #1A3A4A, caption Inter 14px #6B7280, rounded-2xl shadow-sm hover shadow-md, dot active #8B5A2B inactive #E5E7EB.

### Map Leaflet + Nearby

| Library | Tradeoff | Edge | Choice |
|---------|----------|------|--------|
| Leaflet + OpenStreetMap | Free, no API key, custom pin #8B5A2B, 2h setup | Map fail → fallback static image + address text + WhatsApp link | Leaflet, not Google (no key) |

**Nearby:** Beach 0.5km, Cafe 0.2km, Spa 0.3km in card cream #FDF6EC, pin #8B5A2B.

### Loyalty Points

| Rule | Value | Table | Edge |
|------|-------|-------|------|
| Earn | 10 points per night, only checked_out | users.loyalty_points INT, loyalty_tx (user_id, points, booking_id) | Cancelled/rejected gives 0, ACID |
| Redeem | 100 points = IDR 100k discount (auto apply or code LOYALTY100) | min 100 points, quota none | Not enough points → 400 Requires 100 points, you have 80 |
| Badge | Gold #C9A86A in navbar, progress 80/100 to free night | Warm premium, not green | Display 0 if guest |

## Bucket B Grill

### Occupancy Calendar Drag

| Library | Tradeoff | Edge | Choice |
|---------|----------|------|--------|
| FullCalendar (Vue wrapper) + WarmAura | Most elegant for 7-day view, drag booking to shift dates, 2 days work | Drag overlap → recalc FOR UPDATE available, if full → toast + revert | FullCalendar 7-day view, not 30 (light) |

**Colors:** available cream #FDF6EC, occupied #8B5A2B, dirty orange #EF6C00, maintenance grey #6B7280, rounded.

### Housekeeping Kanban

| Library | Columns | Tradeoff | Edge |
|---------|---------|----------|------|
| PrimeVue Card + vuedraggable | 4 cols Available/Occupied/Dirty/Maintenance drag STD-101 | 1 day, wow kanban, not select dropdown | Drag to Occupied without booking → 409 Only checked_in bookings can occupy |

**Hierarchy:** Column header #1A3A4A 14px uppercase tracking, card #8B5A2B border when occupied.

### Add-ons

| Add-on | Price | Table | Edge |
|--------|-------|-------|------|
| Breakfast | 50k | addons + booking_addons M:N | Only pending/waiting can add, total = room + addons |
| Airport transfer | 150k | | |
| Extra bed | 100k | | |

**Total:** room total_price = (price* nights * weekend) + sum addons - voucher - loyalty.

### Dynamic Pricing Weekend +20%

| Rule | Formula | Edge |
|------|---------|------|
| Weekend | Fri,Sat price*1.2 | isWeekend(check_in or check_out) → +20%, voucher calculated from weekend total |

**Choice:** Simple weekend only, not seasonal complex (YAGNI).

## Bucket C Grill

### Realtime WebSocket Admin

| Stack | Tradeoff | Edge | Choice |
|-------|----------|------|--------|
| Fiber websocket `/ws/admin` + Vue useWebSocket | 1.5 days, admin gets New booking #46 toast without refresh | WS disconnect → fallback polling 30s, never blocks booking | Fiber WS admin only, not customer |

### Full-text Search ILIKE

| Query | Tradeoff | Edge |
|-------|----------|------|
| `WHERE name ILIKE %q OR description ILIKE %q` | 1h, YAGNI, fast for 4 types | Empty q → all, 0 result → No rooms match |

**Choice:** ILIKE, not tsvector (overkill for 4).

### Room Photo Upload (5 per type)

| Path | Validation | Table | Edge |
|------|------------|-------|------|
| `storage/room_types/{id}/` uuid.jpg | jpg/png 5MB, max 5 | room_type_images (room_type_id, url) | 6th → 400, php → 400 |

**Dependency:** Gallery uses this, so upload first.

### Export Reports CSV

| Lib | Tradeoff | Edge |
|-----|----------|------|
| encoding/csv Go | 0.5h | No data → No data for selected dates |

**Choice:** CSV first, PDF already via invoice.

### Rate Limit 60/min

| Lib | Edge |
|-----|------|
| fiber/limiter | 429 Too many requests, try in 60s |

## Dependencies & Order

```
Upload Foto (C) → Gallery (A)
Peta standalone
Loyalty → after checked_out (exists)
Calendar (B) → bookings + room_units (exists)
Housekeeping (B) → room_units status (exists)
Add-ons (B) → booking total (exists)
Dynamic Pricing (B) → price calc (exists)
Realtime (C) → bookings create/verify (exists)
Search (C) → standalone
```

**YOLO Order:** Upload → Gallery/Map → Loyalty → Calendar → Housekeeping → Add-ons → Dynamic Pricing → Realtime → Search → Export → Rate Limit.

## Visual Hierarchy Premium WarmAura

- Title Playfair 28px #1A3A4A + gold line 1px #C9A86A
- Body Inter 16px #6B7280
- Cards rounded-2xl shadow-sm hover shadow-md, spacing xl 48px
- Contrast 4.5:1, grid 1 col mobile / 2 tablet / 3-4 desktop
- Log-only email, no SMTP, no em dashes

## Definition of Done Enrich

- [ ] Gallery 5 photos carousel + lightbox + placeholder
- [ ] Map Leaflet pin #8B5A2B + nearby cards
- [ ] Loyalty 10/night 100=100k badge gold progress
- [ ] Calendar 7 days drag colors FOR UPDATE
- [ ] Kanban 4 cols drag
- [ ] Add-ons 3 items total calc
- [ ] Weekend +20%
- [ ] WS admin new booking toast fallback polling
- [ ] Search ILIKE
- [ ] Upload 5 photos admin
- [ ] Export CSV + Rate limit 429
- [ ] All builds ok, beauty hierarchy

> Locked, then YOLO no git ops until push. Next docs/15-fase6.md will summarize build.
