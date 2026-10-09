# 16 — Enrich Extra Grill: Chat + Dark Mode + Multi-language + Interactive Map (LOCKED 2026-10-09)

> Grill 1,2,3,4 before YOLO. Premium WarmAura hierarchy #8B5A2B, beautiful proper, no git ops until push.

## Decisions Locked (1,2,3,4)

| # | Feature | Stack Choice | Why Premium | No Build |
|---|---------|--------------|-------------|----------|
| **1 Chat Support** | WA Widget + In-app Chat WS | In-app chat `messages` table + WS `/ws/chat` broadcast, WA widget float button `https://wa.me/` + bell inbox admin | Hotel feels alive, guest asks balcony, admin replies realtime, WS reuse hub | 1.5 days, no SMTP |
| **2 Dark Mode Premium** | Tailwind `dark:` class toggle + `localStorage` + `prefers-color-scheme` | Cream `#FDF6EC` → Deep `#1A3A4A` bg, card `#2D3748` border `#4A5568`, gold `#C9A86A` stays pop, toggle 🌙 in nav rounded-full | Night beauty, portfolio unique, all hotel light only | 1 day, no extra lib |
| **3 Multi-language EN/ID** | `vue-i18n` 9 + JSON `en.json` `id.json` toggle 🇬🇧🇮🇩 in nav → `localStorage` | Home `Book Your Stay` ↔ `Booking Sekarang`, proper international | EN for recruiter, ID for local | 1 day, 80 keys |
| **4 Interactive Map Booking** | Leaflet 1.9 + custom pin #8B5A2B per room type at Ubud offsets, click pin → popup Card + Reserve | Guest picks room from map, not list, spatial wow, warm pin custom | 1 day, free OSM |  |

## Grill Details

### 1 Chat Support (WA Widget + In-app)

| Aspect | Choice | Tradeoff | Edge |
|--------|--------|----------|------|
| WA Widget | Float button bottom-right #25D366 WhatsApp green, href wa.me/628..., 48x48 rounded-full, hover scale | Free, no backend, customer → WA app | Number not set → fallback `+6281234567890` config |
| In-app Chat | Table `messages` (id, user_id, booking_id nullable, message TEXT, is_admin BOOL, created_at) + `GET/POST /api/chat` + WS `/ws/chat` broadcast | 1.5 days, realtime, admin inbox DataTable per customer | Empty message → 400 Enter a message, max 500 chars, only auth, customer can chat only own, admin all |
| Realtime | Reuse hub `/ws/admin` or new `/ws/chat` → broadcast `{type:chat_message, message}` | WS reuse, fallback polling not needed for chat | WS disconnect → toast Reconnecting |

**Hierarchy:** Bubble customer #8B5A2B cream text right, admin #E5E7EB left, rounded-2xl, spacing md, InputText + Send #8B5A2B.

### 2 Dark Mode Premium

| Aspect | Choice | Tradeoff | Edge |
|--------|--------|----------|------|
| Toggle | Sun/Moon icon lucide Sun/Moon in nav, click → `document.documentElement.classList.toggle('dark')` + `localStorage theme` + `prefers-color-scheme` media | No lib, Tailwind dark: works | Flash of light → set early in main.js before mount |
| Colors | Light cream #FDF6EC → dark bg #1A3A4A, card #2D3748 border #4A5568, text cream, gold #C9A86A accent unchanged, primary #8B5A2B hover #6F4620 | Gold pop in dark, not washed | Contrast 4.5:1 both modes, WCAG AA |
| Coverage | App.vue, HomeView hero, cards, AdminView | 1 day sweeping | Chart colors adapt via computed dark |

**WarmAura Dark:** Header deep #1A3A4A gold line, card dark #2D3748, not pure black #000.

### 3 Multi-language EN/ID

| Library | Keys | Toggle | Tradeoff |
|---------|------|--------|----------|
| vue-i18n 9 | 80 keys (nav, hero, search, cards, booking, wishlist, admin, chat) | 🇬🇧🇮🇩 flags in nav, localStorage `locale`, fallback en | 1 day, proper intl |

**Keys:** `nav.rooms` Rooms/Kamar, `hero.title` Your warm home / Rumah hangat Anda, `cta.book` Book Your Stay / Booking Sekarang, `search.guests` Guests / Tamu etc.

### 4 Interactive Map Booking

| Library | Pin | Popup | Tradeoff |
|---------|-----|-------|----------|
| Leaflet 1.9 CDN | Custom divIcon #8B5A2B per room type at Ubud offsets (Standard -8.519,115.263 ±0.002) | Click pin → popup Card Deluxe Rp 550k + Reserve This Room button → scroll to card | Free OSM, no API key, fallback static |

**Interaction:** Map below gallery, height 400px rounded-2xl, pins 4, popup warm cream, Reserve scrolls and preselects room.

## Dependencies & YOLO Order

```
Dark Mode toggle (standalone)
Multi-language i18n (standalone)
Leaflet map pins (standalone)
Chat WS + messages table (after migrations)
```

**Order:** Migrations messages table → Dark Mode → i18n → Leaflet map + pins → Chat widget + inbox → wiring WS.

## Visual Hierarchy

- Title Playfair 28px #1A3A4A (dark cream) + gold line #C9A86A
- Card rounded-2xl shadow-sm hover shadow-md, spacing xl 48px
- No em dashes, no bullshit, English base + ID toggle

## DoD

- [ ] WA widget float 48px #25D366
- [ ] In-app chat messages table WS broadcast inbox
- [ ] Dark mode toggle 🌙 cream↔deep gold pop contrast 4.5:1 no flash
- [ ] i18n 80 keys toggle persist
- [ ] Leaflet 4 pins popup Reserve scroll
- [ ] All builds ok no git ops until push
