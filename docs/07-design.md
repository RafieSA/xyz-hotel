# 07 — Design Guidelines (UI/UX Manual untuk Agentic AI Coding)

> Panduan ini WAJIB dibaca setiap agent sebelum ngoding UI. Tujuannya: hasil UI konsisten, tidak ngawur, dan mudah dipahami pemula.

## 1. Brand & Tema Hotel

**Konsep:** Hotel bintang 3 modern, hangat, premium tapi membumi — bukan mewah berlebihan, bukan minimalis kaku. Target: keluarga & traveler yang cari nyaman.

**Inspirasi visual:** Kayu hangat + putih bersih + aksen emas — seperti hotel di Bali/Ubud yang calm & trustworthy.

## 2. Color Palette (Hex Code Locked)

| Token | Hex | Penggunaan | Contoh |
|-------|-----|------------|--------|
| **Primary** | `#8B5A2B` (Warm Brown) | Tombol utama, header, CTA "Booking Sekarang" | `bg-[#8B5A2B]` |
| **Primary Hover** | `#6F4620` | Hover tombol primary | |
| **Primary Light** | `#FDF6EC` (Cream) | Background section, card highlight | |
| **Secondary** | `#1A3A4A` (Deep Teal) | Navbar, footer, teks judul | |
| **Accent Gold** | `#C9A86A` | Badge premium, harga, rating star, garis dekoratif | |
| **Success** | `#2E7D32` | Status `verified`, `checked_in` | |
| **Warning** | `#EF6C00` | `pending_payment`, `waiting_verification` | |
| **Danger** | `#C62828` | `rejected`, `expired`, `cancelled` | |
| **Neutral 900** | `#1F2937` | Teks utama | |
| **Neutral 500** | `#6B7280` | Teks sekunder, placeholder | |
| **Neutral 200** | `#E5E7EB` | Border, divider | |
| **Neutral 50** | `#F9FAFB` | Background page | |

**Aturan:**
- Jangan pakai biru `#0A369D` (itu untuk xyz-haircut) — xyz-hotel pakai coklat hangat.
- Rasio kontras minimal 4.5:1 untuk teks (WCAG AA).
- 1 halaman maksimal 3 warna dominan: Primary + Secondary + Neutral.

## 3. Typography

| Level | Font | Size (Desktop) | Weight | Penggunaan |
|-------|------|----------------|--------|------------|
| H1 | `Playfair Display` / `Poppins` | 36-42px | 700 | Judul landing hero |
| H2 | `Poppins` | 28px | 600 | Judul section |
| H3 | `Poppins` | 20px | 600 | Nama tipe kamar |
| Body | `Inter` | 16px | 400 | Deskripsi, form label |
| Small | `Inter` | 14px | 400 | Caption, helper |
| Price | `Poppins` | 18px | 700 | Harga/malam |

**Aturan:**
- Jangan pakai >2 font family. `Poppins` untuk judul, `Inter` untuk body.
- Line height 1.5 untuk body, 1.2 untuk heading.
- Fallback: `system-ui, sans-serif`.

## 4. Spacing & Radius

| Token | Nilai | Penggunaan |
|-------|-------|------------|
| `space-xs` | 4px | Gap ikon-teks |
| `space-sm` | 8px | Padding tombol kecil |
| `space-md` | 16px | Padding card |
| `space-lg` | 24px | Gap antar section |
| `space-xl` | 48px | Padding section |
| `radius-sm` | 8px | Input, badge |
| `radius-md` | 12px | Card kamar |
| `radius-lg` | 16px | Modal, hero image |
| `radius-full` | 9999px | Avatar, pill badge |

## 5. Komponen UI (Aturan Agentic Coding)

### Button
- Primary: `bg-[#8B5A2B] text-white hover:bg-[#6F4620] radius-md px-6 py-3`
- Secondary: `border-2 border-[#8B5A2B] text-[#8B5A2B] bg-white`
- Disabled: `opacity-50 cursor-not-allowed`
- Selalu ada `loading` state (spinner) saat submit booking/upload.

### Card Kamar
```
┌─────────────────────┐
│ [Foto 16:9]         │
│ Deluxe · Rp 550rb/malam │
│ Kapasitas 2 orang · Balkon │
│ [Lihat Detail] [Booking] │
└─────────────────────┘
```
- Shadow `sm`, hover `shadow-md` + `translate-y-[-2px]` (halus 200ms).
- Selalu tampilkan harga, kapasitas, 1 fasilitas unggulan.

### Status Badge (Booking)
| Status | Warna |
|--------|-------|
| `pending_payment` | `bg-orange-100 text-[#EF6C00]` |
| `waiting_verification` | `bg-yellow-100 text-[#F9A825]` |
| `verified` | `bg-green-100 text-[#2E7D32]` |
| `checked_in` | `bg-teal-100 text-[#00695C]` |
| `checked_out` | `bg-gray-100 text-[#616161]` |
| `rejected/expired/cancelled` | `bg-red-100 text-[#C62828]` |

### Form
- Label di atas input, `*` untuk required, helper di bawah.
- Error: border `red-500` + teks merah 14px + ikon ⚠️.
- Success: border `green-500` + checkmark.

### Kalender Availability
- Tanggal kosong = putih, terisi penuh = abu-abu strikethrough + tooltip "Penuh".
- Range dipilih = `bg-[#FDF6EC] border-[#8B5A2B]`.
- Selalu tampilkan harga/malam di bawah kalender.

## 6. Layout

**Frontoffice:**
- Navbar sticky: Logo kiri, menu tengah (Kamar, Fasilitas, Kontak), CTA Booking kanan.
- Hero: Foto hotel full-width + overlay gelap 40% + search box (check-in/out + tamu) di tengah.
- Footer: Alamat, kontak WA, peta, sosmed.

**Backoffice:**
- Sidebar kiri: Dashboard, Kamar, Booking, Voucher, Laporan, User.
- Topbar: Nama user + role badge + logout.
- Content: Tabel + filter + pagination (jangan infinite scroll untuk data operasional).

## 7. Responsive (Mobile First)

| Breakpoint | Lebar | Aturan |
|------------|-------|--------|
| Mobile | <640px | 1 kolom card, search box vertikal, hamburger menu |
| Tablet | 640-1024px | 2 kolom card |
| Desktop | >1024px | 3 kolom card, sidebar backoffice |

- Tap target minimal 44×44px (Apple HIG).
- Jangan pakai hover sebagai satu-satunya interaksi di mobile.

## 8. Aksesibilitas (Wajib)
- Semua gambar punya `alt`.
- Kontras teks vs background ≥4.5:1 (cek via `07-design.md` palette).
- Keyboard navigable (Tab, Enter, Esc untuk modal).
- `aria-label` untuk ikon tanpa teks.

## 9. Contoh Implementasi (Tailwind)

```html
<!-- Tombol Primary -->
<button class="bg-[#8B5A2B] hover:bg-[#6F4620] text-white font-semibold px-6 py-3 rounded-xl transition-colors disabled:opacity-50">
  Booking Sekarang
</button>

<!-- Card Kamar -->
<div class="bg-white rounded-xl shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all overflow-hidden">
  <img src="deluxe.jpg" alt="Kamar Deluxe" class="w-full aspect-[16/9] object-cover" />
  <div class="p-4 space-y-2">
    <h3 class="font-semibold text-[#1A3A4A]">Deluxe</h3>
    <p class="text-sm text-[#6B7280]">Kapasitas 2 · Balkon · Mini fridge</p>
    <p class="font-bold text-[#8B5A2B]">Rp 550.000 <span class="font-normal text-sm">/ malam</span></p>
  </div>
</div>
```

## 10. Yang Dilarang
- ❌ Jangan ganti primary color tanpa diskusi (sudah lock `#8B5A2B`).
- ❌ Jangan pakai animasi berlebihan (max 200-300ms, ease-out).
- ❌ Jangan handwritten CSS dari 0 kalau Tailwind sudah ada.
- ❌ Jangan pakai warna neon / gradient norak.

## 11. Referensi untuk Agent
- Sebelum ngoding UI, baca file ini + `02-prd.md` + `05-booking-flow...md`.
- Pakai Tailwind CSS (jika stack Laravel + Inertia Vue, sudah include).
- Simpan token di `tailwind.config.js` atau `resources/css/app.css` sebagai CSS variables.

> Setelah stack lock (Q1), token di atas akan di-sync ke `tailwind.config.js` biar tidak hardcode hex di tiap file.

## 12. Definition of Done Design
- [ ] Semua halaman pakai palette di atas (cek via Figma/Chrome pick).
- [ ] Mobile & desktop tidak pecah.
- [ ] Status badge konsisten.
- [ ] Tidak ada kontras gagal WCAG.

> Next: seed tipe kamar di `database/seeders` setelah scaffolding (4 tipe locked).
