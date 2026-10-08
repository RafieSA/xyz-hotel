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

## Rekomendasi Final
**Pilih Opsi A: Laravel 12 + Inertia + Vue 3 + Postgres (atau MySQL).**

**Alasan presisi (bukan asumsi):**
1. **Faktual:** PHP 8.2, Postgres 18.4, MySQL 9.6 sudah ada di M4 kamu — verifikasi via `bash`. Tanpa Docker, Laravel jalan dengan `php artisan serve`.
2. **Portfolio:** Hotel dengan Laravel + Vue adalah showcase yang mudah dipahami recruiter/client Indonesia (banyak pakai PHP).
3. **Fitur level lengkap:** Voucher, laporan, upload bukti, RBAC 4 role — Laravel sudah punya paket matang.
4. **Security:** Bawaan Laravel cegah IDOR/BOLA via Policy, SQLi via Eloquent parameterized, CSRF token otomatis.
5. **YAGNI & DRY:** Satu repo, satu bahasa backend, tidak perlu pisah frontend/backend repo di Fase 0.
6. **No handwritten from scratch:** Pakai Breeze/Inertia starter (resmi Laravel) — bukan ngoding auth dari 0.

**Alternatif jika kamu big no Laravel juga:** Pilih **Opsi B (Nuxt 3)** — beri tahu, aku ganti rekomendasi.

## Tradeoff yang Harus Kamu Tahu (Jujur)
| Jika Pilih Laravel | Jika Pilih Nuxt/SvelteKit |
|--------------------|----------------------------|
| Deploy gampang di cPanel/VPS PHP | Deploy butuh Node server (lebih mahal) |
| Auth & upload sudah jadi | Auth & upload bikin manual |
| SEO SSR via Inertia (cukup bagus, tapi tidak se-SEO Next/Nuxt murni) | SEO SSR lebih optimal |
| Belajar Blade/Vue | Full JS |

## Pertanyaan Grill untuk Kamu (Jawab 1 saja)
**"Setuju Opsi A (Laravel + Inertia Vue)? Atau kamu mau Opsi B (Nuxt) / Opsi C (SvelteKit)?"**

> Jawab singkat: "A" / "B" / "C" / "Ada opsi lain". Setelah itu kita lock dan lanjut grill 4 role & flow. Tidak akan lanjut scaffolding sebelum kamu lock.

## Best Practice yang Akan Kita Terapkan (Apapun Stack)
- Clean Code, SRP (1 controller 1 tanggung jawab), Error Handling + Logging (bukan `try-catch` kosong).
- Validasi di FormRequest (Laravel) / Zod (JS) — jangan percaya input user.
- Transaction untuk booking (ACID).
- Testable: minimal unit test untuk availability logic.

> Setelah lock, file ini akan di-update dengan "DECISION: Opsi X — Locked 2026-10-08".
