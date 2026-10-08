# 06 — Context & Constraints (Portfolio, Local, Security)

> File ini menjawab Q5: "Lo yang atur, bebas sekreatif mungkin" — tapi tetap terukur.

## Konteks Portfolio Pribadi
- **No deadline, no client, no bos** — kamu berdiri 2 kaki sendiri. Artinya: kualitas > kecepatan. Kita tidak kejar sprint, kita kejar showcase yang bisa kamu banggakan di CV/GitHub.
- **Bebas kreatif:** Tipe kamar, harga, fasilitas, desain — agent atur bebas tapi tetap masuk akal bisnis hotel bintang 3.
- **Tipe kamar — LOCKED 2026-10-08 (akan di-seed via `database/seeders/RoomSeeder.php`):**

| Tipe | Kapasitas | Kasur | Fasilitas | Harga/malam | Unit |
|------|-----------|-------|-----------|-------------|------|
| Standard | 2 | 1 Queen | AC, TV, Kamar mandi dalam | Rp 350.000 | 8 |
| Deluxe | 2 | 1 Queen + Sofa | + Balkon, Mini fridge | Rp 550.000 | 5 |
| Family | 4 | 2 Queen | + Dapur mini, 2 kamar | Rp 850.000 | 3 |
| Suite | 2 | 1 King | + Living room, Jacuzzi | Rp 1.250.000 | 2 |

> Total 18 unit fisik (STD-101..108, DLX-201..205, FAM-301..303, STE-401..402). Harga di-snapshot ke `bookings.total_price` saat booking.

## Prinsip yang Wajib Dijaga (No Future Debt)
| Prinsip | Contoh Penerapan di xyz-hotel |
|---------|-------------------------------|
| **YAGNI** | Jangan bikin multi-hotel, multi-kamar per booking, atau payment gateway di v1. |
| **DRY** | Logic availability di 1 service `AvailabilityService`, jangan copy-paste di controller. |
| **KISS** | Validasi tanggal simple: `check_out > check_in`, jangan bikin engine pricing kompleks dulu. |
| **SRP** | `BookingController` hanya HTTP, `BookingService` handle transaction, `BookingPolicy` handle izin. |
| **SOLID** | Dependency injection untuk service, bukan `new` di controller. |
| **ACID** | Booking creation dalam 1 transaction. |
| **Clean Code** | Nama variabel `occupiedUnits` bukan `x`, fungsi < 30 baris. |
| **Testable** | Availability logic bisa di-unit-test tanpa HTTP. |

## Security — Haram Bocor (Checklist)

| Ancaman | Cara Cegah di xyz-hotel |
|---------|--------------------------|
| **SQL Injection** | Eloquent ORM + parameterized query, jangan raw string concat |
| **IDOR / BOLA** | Policy cek `booking.user_id == auth.id` |
| **BFLA** | Middleware role di setiap route group |
| **Broken Auth** | Hash argon2/bcrypt, httpOnly cookie, rate limit 5x/menit login |
| **XSS** | Escape output di Vue (`{{ }}` otomatis escaped), validasi input |
| **CSRF** | Laravel CSRF token otomatis untuk Inertia |
| **Sensitive Data** | Jangan log password, jangan expose `user.email` ke customer lain |
| **Security Misconfig** | `.env` tidak commit, `APP_DEBUG=false` di prod, error message generic |
| **SSRF** | Validasi URL upload (hanya local storage, bukan fetch URL eksternal) |

## Scalability & Maintainability (Portfolio Harus Tahan Lama)
- **Pagination** di daftar booking/kamar (jangan `SELECT *` tanpa limit).
- **Index DB:** `bookings(check_in, check_out, room_type_id, status)`, `users(email)`.
- **Soft delete** untuk kamar & user (jangan hard delete — histori booking butuh).
- **Audit log** tabel terpisah.
- **Dokumentasi API** di `docs/02-api.md` (akan dibuat Fase 1).
- **Logging** terstruktur: `Log::info('booking.created', ['id'=>..])`.

## Kreativitas yang Diizinkan (Karena No Rules)
- Desain landing page bebas — bisa pakai tema warm hotel (gold, cream) beda dari xyz-haircut biru.
- Fitur bonus portfolio: peta lokasi hotel, galeri foto, FAQ, kontak WhatsApp.
- Laporan bisa pakai chart (Chart.js) — showcase skill.

## Definition of Done Global (Portfolio Grade)
- [ ] Kode rapi, ada README cara jalan local
- [ ] Tidak ada TODO/FIXME sisa
- [ ] Screenshots / demo GIF di README
- [ ] Repo GitHub public, commit history bersih
- [ ] Tidak ada secret ter-commit

## Next Step Setelah Fase 0
1. Lock tech stack (jawab grill `03`)
2. Lock role & flow (jawab grill `04` & `05`)
3. `git remote add` + push ke GitHub (kamu minta nanti)
4. Fase 1: scaffolding Laravel + migrasi DB

> Kalau kamu setuju tipe kamar di atas, bilang "setuju tipe kamar". Kalau mau ubah, bilang maunya gimana.
