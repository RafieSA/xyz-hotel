# 03 — Keputusan Tech Stack (Tanpa Next.js)

> Grill Q1 — Dijawab jujur & presisi.

## Kenapa Sebelumnya Rekomendasi Next.js?
Jujur, karena 3 alasan (bukan fanboy):

| Alasan | Penjelasan Membumi |
|--------|---------------------|
| SEO hotel butuh SSR | Hotel butuh Google nemu halaman kamarnya. Next.js jago SSR. Tapi **bukan satu-satunya** — Laravel SSR, Nuxt, SvelteKit juga bisa. |
| Ekosistem besar | Banyak template hotel Next.js siap pakai. Tapi template Laravel & Nuxt juga banyak. |
| Referensi umum | Kebanyakan portfolio hotel di JS pakai Next. Tapi kamu **big no**, jadi kita hormati — preferensi kamu nomor 1. |

**Kesimpulan:** Next.js bukan wajib. Kita bisa dapat hasil sama bagus tanpa Next.js. Haram memaksa stack yang kamu tidak suka.

## Opsi Pengganti (Tanpa Next.js) — Tradeoff Jujur

### Opsi A — Laravel 12 + Inertia + Vue 3 + MySQL/Postgres ⭐ REKOMENDASI
**Gambaran:** Satu codebase Laravel jadi backend + frontend (Vue di dalam Laravel via Inertia). Seperti `xyz-haircut` tapi versi PHP.

| Aspek | Nilai |
|-------|-------|
| **Kelebihan** | PHP 8.2 kamu sudah siap, MySQL/Postgres siap, **tanpa Docker**. Laravel punya auth, CSRF, ORM aman (anti-SQLi), RBAC (Gates/Policies), queue, mail, storage upload bukti — semua built-in. Inertia bikin UX SPA tanpa ribet API terpisah. Cocok untuk portfolio hotel (banyak hotel beneran pakai PHP). Local di M4 ngebut. |
| **Kekurangan** | Butuh belajar Inertia kalau belum pernah. Bukan JS murni. |
| **Scalability** | Tinggi — Laravel scale ke multi-hotel gampang. |
| **Security** | Bawaan Laravel sudah cegah SQLi, XSS, CSRF, IDOR via Policy. |
| **Contoh** | `app/Http/Controllers/BookingController.php` handle availability dengan DB transaction. `resources/js/Pages/Booking/Create.vue` tampilkan kalender. |

**Cocok jika:** Kamu mau cepat jadi, local-first, portfolio yang mudah di-deploy ke shared hosting/VPS murah.

### Opsi B — Nuxt 3 (Vue) + Nitro + Postgres (via Prisma/Drizzle)
**Gambaran:** Sepupu Next.js tapi versi Vue. SSR juga jago.

| Aspek | Nilai |
|-------|-------|
| **Kelebihan** | Vue lebih membumi dari React, SSR oke untuk SEO hotel, komunitas besar. |
| **Kekurangan** | Butuh setup auth manual (tidak sekomplit Laravel). Butuh backend terpisah atau Nitro. Di M4 tanpa Docker tetap jalan, tapi deploy butuh Node server. |
| **Scalability** | Tinggi. |
| **Security** | Harus handle CSRF/SQLi manual. |

**Cocok jika:** Kamu cinta Vue dan mau full JS.

### Opsi C — SvelteKit + Postgres
**Gambaran:** Paling ringan & cepat, syntax simple.

| Aspek | Nilai |
|-------|-------|
| **Kelebihan** | Ringan, cepat, bundle kecil, SSR bagus. |
| **Kekurangan** | Ekosistem hotel template lebih sedikit, community lebih kecil dari Laravel/Nuxt. |
| **Scalability** | Tinggi tapi butuh setup manual. |

### Opsi D — Dart Frog + Flutter Web (Konsisten xyz-haircut)
**Gambaran:** Samakan dengan project sebelumnya.

| Aspek | Nilai |
|-------|-------|
| **Kelebihan** | Konsisten, kamu sudah paham. |
| **Kekurangan** | Flutter Web SEO lemah (hotel butuh SEO), bundle besar, kurang template hotel. **Tidak direkomendasikan untuk website hotel.** |

## ✅ DECISION LOCKED — 2026-10-08 — Backend Golang + Frontend Vue.js (G1)

**Pilihan final Rafie: Backend Golang (Go 1.26.3) + Frontend Vue 3 + Postgres 18.4.**

### Arsitektur Final
```
[ Vue 3 + Vite + Tailwind #8B5A2B ]  -- REST API JSON (JWT) -->  [ Go API — Fiber/Gin + sqlx ]  --> [ Postgres 18.4 ]
        |                                                       |                              |
   Frontoffice: search, booking, upload                    RBAC 4 role (owner/manager/       tables: users, room_types,
   Backoffice: dashboard, verifikasi, laporan               receptionist/customer)             room_units, bookings, vouchers,
   Design token dari docs/07-design.md                     Transaction FOR UPDATE             audit_logs
                                                           Upload bukti (5MB, jpg/png/pdf)
                                                           Expired 12 jam (ticker)
```

### Detail Stack
| Lapisan | Teknologi | Versi | Alasan |
|---------|-----------|-------|--------|
| Backend | **Go + Fiber** (alternatif Gin) | Go 1.26.3 | Fiber paling ngebut, syntax mirip Express, ringan di M4 |
| DB | **Postgres** | 18.4 | ACID + `FOR UPDATE` solid, sudah ada lokal |
| Frontend | **Vue 3 + Vite + Tailwind** | Node 26 | Vue membumi, Vite super cepat, Tailwind untuk design #8B5A2B |
| Auth | JWT access 15 menit + refresh (httpOnly) + bcrypt | — | Cegah Broken Auth |
| Upload | `storage/uploads` lokal, validasi MIME | — | Manual transfer bukti |
| Repo | Monorepo `backend/` + `frontend/` + `docs/` | — | Local `main` branch, no worktrees |

### Kenapa Go + Vue Cocok untuk xyz-hotel
1. **Faktual M4:** Go 1.26.3 & Node 26 & Postgres sudah ready — verifikasi `go version` 2026-10-08.
2. **Portfolio Go:** Langka & bernilai tinggi — showcase "bisa Go + concurrency" beda dari 100 pelamar Laravel.
3. **Level lengkap tetap bisa:** Voucher, laporan, review — semua bisa di Go, cuma lebih manual (no magic ORM).
4. **Security:** Harus disiplin pakai placeholder `$1,$2` (anti-SQLi), middleware JWT + role (anti IDOR/BOLA/BFLA), Vue auto-escape (anti-XSS).
5. **No handwritten from scratch yang sia-sia:** Pakai Fiber (resmi), `sqlx`, `golang-migrate`, `go-playground/validator` — jangan bikin router dari 0.

### Tradeoff yang Disepakati (Jujur)
| Go + Vue | Konsekuensi |
|----------|-------------|
| Dev lebih lama 2-3x vs Laravel | Auth/RBAC/upload/validasi tulis manual — tapi no deadline, jadi oke |
| Pisah backend/frontend | Setup 2x (`go run` + `npm run dev`), tapi clean separation |
| Template hotel Go sedikit | Frontend Vue bikin dari 0 pakai design.md |
| Performa paling ngebut | Handle race condition booking dengan `FOR UPDATE` tetap ACID |

### Best Practice yang Wajib (Go + Vue)
- Clean Code, SRP: `handler` hanya HTTP, `service` handle transaction, `middleware` handle auth/role.
- Validasi: `validator` di Go + `Zod` di Vue — jangan percaya input user.
- Transaction: `BEGIN; SELECT ... FOR UPDATE; INSERT; COMMIT;` untuk booking.
- Logging: `log/slog` terstruktur + `AuditLog` tabel.
- Testable: `AvailabilityService` bisa di-unit-test tanpa HTTP.

> Grill Q1 locked. Q3/Q5/Q6 sudah locked sebelumnya. Siap Fase 1 scaffolding.
