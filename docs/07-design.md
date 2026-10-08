# 07 — Design Guidelines (UI/UX Manual for Agentic AI Coding)

> Every agent MUST read this guide before coding UI. Goal: consistent, well-crafted, beginner-friendly results.

## 1. Brand & Hotel Theme

**Concept:** A modern 3-star hotel — warm, premium yet grounded. Not overly luxurious, not rigidly minimal. Target audience: families & travelers seeking comfort.

**Visual inspiration:** Warm wood + clean white + gold accents — like a calm & trustworthy hotel in Bali/Ubud.

## 2. Color Palette (Hex Codes Locked)

| Token | Hex | Usage | Example |
|-------|-----|-------|---------|
| **Primary** | `#8B5A2B` (Warm Brown) | Primary buttons, header, "Book Now" CTA | `bg-[#8B5A2B]` |
| **Primary Hover** | `#6F4620` | Primary button hover | |
| **Primary Light** | `#FDF6EC` (Cream) | Section background, card highlight | |
| **Secondary** | `#1A3A4A` (Deep Teal) | Navbar, footer, heading text | |
| **Accent Gold** | `#C9A86A` | Premium badge, price, rating stars, decorative lines | |
| **Success** | `#2E7D32` | Status `verified`, `checked_in` | |
| **Warning** | `#EF6C00` | `pending_payment`, `waiting_verification` | |
| **Danger** | `#C62828` | `rejected`, `expired`, `cancelled` | |
| **Neutral 900** | `#1F2937` | Primary text | |
| **Neutral 500** | `#6B7280` | Secondary text, placeholder | |
| **Neutral 200** | `#E5E7EB` | Border, divider | |
| **Neutral 50** | `#F9FAFB` | Page background | |

**Rules:**
- Do not use blue `#0A369D` (that belongs to xyz-haircut) — xyz-hotel uses warm brown.
- Minimum contrast ratio 4.5:1 for text (WCAG AA).
- Maximum 3 dominant colors per page: Primary + Secondary + Neutral.

## 3. Typography

| Level | Font | Size (Desktop) | Weight | Usage |
|-------|------|----------------|--------|-------|
| H1 | `Playfair Display` / `Poppins` | 36-42px | 700 | Landing hero headline |
| H2 | `Poppins` | 28px | 600 | Section headline |
| H3 | `Poppins` | 20px | 600 | Room type name |
| Body | `Inter` | 16px | 400 | Description, form labels |
| Small | `Inter` | 14px | 400 | Caption, helper text |
| Price | `Poppins` | 18px | 700 | Price/night |

**Rules:**
- Do not use more than 2 font families. `Poppins` for headings, `Inter` for body.
- Line height 1.5 for body, 1.2 for headings.
- Fallback: `system-ui, sans-serif`.

## 4. Spacing & Radius

| Token | Value | Usage |
|-------|-------|-------|
| `space-xs` | 4px | Icon-text gap |
| `space-sm` | 8px | Small button padding |
| `space-md` | 16px | Card padding |
| `space-lg` | 24px | Gap between sections |
| `space-xl` | 48px | Section padding |
| `radius-sm` | 8px | Input, badge |
| `radius-md` | 12px | Room card |
| `radius-lg` | 16px | Modal, hero image |
| `radius-full` | 9999px | Avatar, pill badge |

## 5. UI Components (Agentic Coding Rules)

### Button
- Primary: `bg-[#8B5A2B] text-white hover:bg-[#6F4620] radius-md px-6 py-3`
- Secondary: `border-2 border-[#8B5A2B] text-[#8B5A2B] bg-white`
- Disabled: `opacity-50 cursor-not-allowed`
- Always include a `loading` state (spinner) when submitting a booking/upload.

### Room Card
```
┌─────────────────────┐
│ [Photo 16:9]        │
│ Deluxe · Rp 550k/night │
│ Capacity 2 · Balcony │
│ [View Details] [Book] │
└─────────────────────┘
```
- Shadow `sm`, hover `shadow-md` + `translate-y-[-2px]` (smooth 200ms).
- Always show price, capacity, and 1 featured amenity.

### Status Badge (Booking)
| Status | Color |
|--------|-------|
| `pending_payment` | `bg-orange-100 text-[#EF6C00]` |
| `waiting_verification` | `bg-yellow-100 text-[#F9A825]` |
| `verified` | `bg-green-100 text-[#2E7D32]` |
| `checked_in` | `bg-teal-100 text-[#00695C]` |
| `checked_out` | `bg-gray-100 text-[#616161]` |
| `rejected/expired/cancelled` | `bg-red-100 text-[#C62828]` |

### Form
- Label above input, `*` for required fields, helper text below.
- Error: border `red-500` + red text 14px + ⚠️ icon.
- Success: border `green-500` + checkmark.

### Availability Calendar
- Available date = white, fully booked = gray strikethrough + tooltip "Fully booked".
- Selected range = `bg-[#FDF6EC] border-[#8B5A2B]`.
- Always show price/night below the calendar.

## 6. Layout

**Frontoffice:**
- Sticky navbar: Logo on the left, menu in the center (Rooms, Amenities, Contact), Book CTA on the right.
- Hero: Full-width hotel photo + 40% dark overlay + search box (check-in/out + guests) centered.
- Footer: Address, WhatsApp contact, map, social media.

**Backoffice:**
- Left sidebar: Dashboard, Rooms, Bookings, Vouchers, Reports, Users.
- Top bar: User name + role badge + logout.
- Content: Table + filters + pagination (do not use infinite scroll for operational data).

## 7. Responsive (Mobile First)

| Breakpoint | Width | Rule |
|------------|-------|------|
| Mobile | <640px | 1-column cards, vertical search box, hamburger menu |
| Tablet | 640-1024px | 2-column cards |
| Desktop | >1024px | 3-column cards, backoffice sidebar |

- Minimum tap target 44×44px (Apple HIG).
- Do not rely on hover as the sole interaction on mobile.

## 8. Accessibility (Required)
- Every image has an `alt` attribute.
- Text vs. background contrast ≥4.5:1 (check via the palette in `07-design.md`).
- Keyboard navigable (Tab, Enter, Esc for modals).
- `aria-label` for icons without text.

## 9. Implementation Example (Tailwind)

```html
<!-- Primary Button -->
<button class="bg-[#8B5A2B] hover:bg-[#6F4620] text-white font-semibold px-6 py-3 rounded-xl transition-colors disabled:opacity-50">
  Book Now
</button>

<!-- Room Card -->
<div class="bg-white rounded-xl shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all overflow-hidden">
  <img src="deluxe.jpg" alt="Deluxe Room" class="w-full aspect-[16/9] object-cover" />
  <div class="p-4 space-y-2">
    <h3 class="font-semibold text-[#1A3A4A]">Deluxe</h3>
    <p class="text-sm text-[#6B7280]">Capacity 2 · Balcony · Mini fridge</p>
    <p class="font-bold text-[#8B5A2B]">Rp 550,000 <span class="font-normal text-sm">/ night</span></p>
  </div>
</div>
```

## 10. Prohibited
- ❌ Do not change the primary color without discussion (locked as `#8B5A2B`).
- ❌ Do not use excessive animations (max 200-300ms, ease-out).
- ❌ Do not write handwritten CSS from scratch when Tailwind is already available.
- ❌ Do not use neon colors / tacky gradients.

## 11. Reference for Agents
- Before coding UI, read this file + `02-prd.md` + `05-booking-flow...md`.
- Use Tailwind CSS (already included if the stack is Laravel + Inertia Vue).
- Store tokens in `tailwind.config.js` or `resources/css/app.css` as CSS variables.

> Once the stack is locked (Q1), the tokens above will be synced to `tailwind.config.js` so hex codes are not hardcoded in every file.

## 12. Design Definition of Done
- [ ] All pages use the palette above (verify via Figma/Chrome color picker).
- [ ] Mobile & desktop layouts do not break.
- [ ] Status badges are consistent.
- [ ] No WCAG contrast failures.

> Next: seed room types in `database/seeders` after scaffolding (4 types locked).
