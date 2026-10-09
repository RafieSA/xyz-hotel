# 17 — Enrich Extra: Chat + Dark Mode + Multi-language + Interactive Map (COMPLETE 2026-10-09 — YOLO)

> Chat WA widget + in-app WS, dark mode premium WarmAura, EN/ID toggle 89 keys, Leaflet 4 pins popup Reserve. All beautiful hierarchy, no git ops until push.

## Summary

| Feature | Status | Evidence |
|---------|--------|----------|
| **Chat WA Widget + In-app** | ✅ | WA float 48px #25D366 wa.me, In-app POST /api/chat HELLO BALCONY → 201 + GET → 1 message, WS /ws/chat hub broadcast |
| **Dark Mode Premium** | ✅ | toggleTheme class LocalStorage + prefers, cream #FDF6EC ↔ deep #1A3A4A card #2D3748 gold #C9A86A pop, Moon/Sun button #8B5A2B |
| **Multi-language EN/ID** | ✅ | vue-i18n 9 89 keys nav.rooms Rooms/Kamar hero.badge Ubud Since 2024 etc, flags 🇬🇧🇮🇩 persist localStorage |
| **Interactive Map** | ✅ | Leaflet 1.9 CDN link leaflet.css, div#map 400px rounded-2xl L.map -8.519,115.263 zoom 13 4 divIcon pins #8B5A2B popup Reserve scrollToRoom |
| **Builds** | ✅ | `go vet 0`, `go test ok`, `npm build` 357KB 494ms 157 handlers |

## Chat

**Migration 008_chat.sql:** messages (id, user_id FK, booking_id nullable FK, message TEXT 1-500, is_admin BOOL, created_at).

**Routes:**
- `POST /api/chat` {message, booking_id?} Auth validate 1-500, customer booking ownership 403
- `GET /api/chat?booking_id?` Auth customer own only, admin all
- `GET /api/admin/chat/messages?user_id?` RBAC owner/manager
- `GET /ws/chat` WS upgrade Auth, hub broadcast chat_message JSON

**WA Widget:** Float bottom-right 48x48 rounded-full #25D366, hover scale, href wa.me/6281234567890 fallback, lucide MessageCircle.

**Live:**
```bash
POST /api/chat {"message":"Hello, do Standard rooms have balcony?"} → 201 {"id":1}
GET /api/chat → [{"message":"Hello, do Standard rooms have balcony?"}]
```

## Dark Mode

**Tailwind:** darkMode class, composable useTheme.js initTheme (localStorage or prefers-color-scheme), no flash via main.js init before mount.

**Colors:** bg cream #FDF6EC ↔ deep #1A3A4A, card #2D3748 border #4A5568, text cream #FDF6EC, gold #C9A86A accent unchanged, primary #8B5A2B hover #6F4620.

**App nav:** Moon/Sun toggle 9x9 rounded-full #8B5A2B border white/20, aria-label.

## Multi-language

**Lib:** vue-i18n 9, src/i18n/index.js createI18n, src/locales/en.json + id.json 89 keys (nav.rooms Rooms/Kamar, hero.title Your warm home/Rumah hangat, cta.book Book Your Stay/Booking Sekarang etc), App flags toggle persist, HomeView/AdminView use useI18n, no em dashes.

## Interactive Map

**Leaflet:** index.html link leaflet.css CDN, HomeView div#map h-[400px] rounded-2xl, L.map at -8.519,115.263 zoom 13 OSM tiles, 4 divIcon pins #8B5A2B per room type offsets Standard -8.519,115.263 Deluxe -8.521,115.265 etc, bindPopup warm card + Reserve This Room button scrolls to room card preselects, fallback static text if fail.

## How to Run

```bash
cd backend && go run ./cmd/server  # :8080 157 handlers
curl -X POST :8080/api/chat -H "Bearer $TOKEN" -d '{"message":"Hello balcony?"}'
cd frontend && npm run dev  # :5173 toggle 🌙/🇬🇧🇮🇩, map 400px, chat bubble, pins
```

> All SRP, $1 safe, WarmAura #8B5A2B, beautiful hierarchy, YOLO no git ops.
