# 02 — PRD (Product Requirements Document) — Level Lengkap

> Status: Draft Fase 0 — akan diperhalus setelah grill stack & role final.

## 1. Ringkasan Produk
Website booking hotel untuk **satu hotel** dengan frontoffice dan backoffice terpisah, pembayaran manual transfer, dan kalender availability real-time.

## 2. User & Role (4 Role)
Lihat detail di `04-roles-permissions.md`. Ringkas:
- **Super Admin / Owner** — atur semua + lihat laporan uang
- **Manager** — kelola kamar, harga, promo, verifikasi booking
- **Receptionist / Front Desk** — check-in/out, ubah status kamar (bersih/kotor/rusak)
- **Customer** — daftar, cari kamar, booking, upload bukti, review

## 3. Fitur Frontoffice (Pelanggan)

### 3.1 Wajib (v1)
- [ ] Landing page: hero, daftar tipe kamar, fasilitas, CTA booking
- [ ] Search availability: pilih `check_in` & `check_out` → lihat kamar kosong real-time
- [ ] Detail kamar: foto, deskripsi, fasilitas, harga/malam, kapasitas
- [ ] Booking flow: pilih kamar → isi tamu (nama, HP, email) → buat booking `pending_payment` → instruksi transfer
- [ ] Upload bukti transfer (gambar/PDF) → status `waiting_verification`
- [ ] Riwayat booking + status tracking (pending, verified, checked_in, checked_out, cancelled, expired)
- [ ] Auth: register/login (email+password), lupa password, profil

### 3.2 Lengkap (v1 tapi bisa bertahap)
- [ ] Review & rating per tipe kamar (hanya yang sudah check-out bisa review)
- [ ] Voucher/kode promo (potongan %, min. malam, expiry)
- [ ] Wishlist / favorit
- [ ] Invoice PDF setelah verified
- [ ] Email notifikasi (booking dibuat, diverifikasi, expired)

### 3.3 Contoh Skenario Membumi
> **Skenario A:** Ani mau liburan 2 malam (10-12 Okt). Dia buka website → pilih tanggal → sistem cek: Deluxe masih 2 unit kosong → Ani booking 1 kamar Deluxe → dapat instruksi transfer BCA 123456 Rp 1.000.000 → upload bukti → admin verifikasi 10 menit → status jadi `verified` → Ani dapat email invoice.

## 4. Fitur Backoffice (Manajemen)

### 4.1 Kamar & Inventory
- [ ] CRUD tipe kamar (nama, deskripsi, kapasitas, harga/malam, jumlah unit)
- [ ] CRUD unit kamar fisik (contoh: Deluxe-101, Deluxe-102) + status: `available`, `occupied`, `dirty`, `maintenance`
- [ ] Kalender occupancy (lihat semua booking per tanggal)
- [ ] Atur harga musiman / weekend (opsional v1)

### 4.2 Booking & Operasional
- [ ] Daftar booking: filter status, tanggal, tipe kamar
- [ ] Verifikasi bukti transfer (approve/reject + alasan)
- [ ] Check-in / Check-out manual + assign unit kamar fisik
- [ ] Cancel & refund manual (catat alasan)
- [ ] Expired otomatis: booking `pending_payment` yang tidak bayar dalam 2 jam → `expired`

### 4.3 Laporan & Lainnya
- [ ] Laporan harian/bulanan: occupancy rate, revenue, booking per tipe
- [ ] Kelola voucher/promo
- [ ] Kelola user & role
- [ ] Log audit (siapa ubah apa, kapan)

## 5. Alur Bisnis Inti (Disederhanakan)
```
Customer                    System                     Admin
   │                          │                          │
   ├─ search(tgl) ───────────▶│                          │
   │◀─ list kamar kosong ─────┤                          │
   ├─ create booking ────────▶│                          │
   │◀─ pending_payment ───────┤                          │
   ├─ upload bukti ──────────▶│                          │
   │                          ├─ waiting_verification ──▶│
   │                          │◀─ verified/rejected ─────┤
   │◀─ notif verified ────────┤                          │
   │                          │                          ├─ check-in ──▶ occupied
   │                          │                          ├─ check-out ─▶ dirty → available
```

Detail flow + edge cases di `05-booking-flow-and-edge-cases.md`.

## 6. Aturan Bisnis Penting
| Aturan | Contoh |
|--------|--------|
| Check-in 14:00, check-out 12:00 | Booking 10-12 Okt = 2 malam, unit bebas jam 12:00 tgl 12 |
| 1 booking = 1 tipe kamar, N malam | Tidak campur tipe dalam 1 booking (YAGNI) |
| Overlap dilarang | Kamar Deluxe total 5 unit, kalau 5 sudah dibooking tgl X, booking ke-6 harus ditolak |
| Harga = harga/malam × malam − diskon voucher | Voucher cek expiry & quota |
| Review hanya setelah check-out | Cegah review palsu |

## 7. Kebutuhan Non-Fungsional
- **Security:** IDOR/BOLA/BFLA blocked, SQLi via parameterized query, XSS escaped, CSRF token, password hash bcrypt/argon2, JWT/cookie httpOnly.
- **Scalability:** Siap multi-hotel nanti (tambah `hotel_id`), pagination, index di `check_in/out`.
- **Maintainability:** Clean Code, SRP, error handling + logging, testable.
- **ACID:** Booking creation harus transaksi DB (cek availability + insert booking atomik).
- **Local-first:** Jalan di M4 tanpa Docker, migrasi terurut.

## 8. Out of Scope v1
- Payment gateway otomatis
- Multi-hotel
- Channel manager (Agoda)
- Mobile app

## 9. Definition of Done per Fitur
- Ada API + UI + validasi + error handling + logging + test minimal 1 happy + 1 edge case.
- Tidak ada happy-path-only — semua edge case di `05-...` harus di-handle.

> Next: `03-tech-stack-decision.md` untuk pilih stack tanpa Next.js.
