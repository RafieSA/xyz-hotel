# 04 — Roles & Permissions (4 Role)

> Grill Q3 — 4 role untuk satu hotel. Dibuat membumi dengan contoh.

## Daftar Role

| # | Role | Siapa di Dunia Nyata | Analogi |
|---|------|----------------------|---------|
| 1 | **Super Admin / Owner** | Pemilik hotel | Bos besar, lihat uang, atur semua |
| 2 | **Manager** | Manajer operasional | Atur kamar, harga, promo, verifikasi booking |
| 3 | **Receptionist** | Front desk / resepsionis | Check-in/out, ubah status kamar bersih/kotor |
| 4 | **Customer** | Tamu / pelanggan | Booking, bayar, review |

> Nama role bisa diganti (misal: Owner → Super Admin). Yang penting izinnya jelas.

## Matriks Izin (Permission Matrix)

| Fitur | Owner | Manager | Receptionist | Customer |
|-------|-------|---------|:------------:|:--------:|
| **Kelola user & role** | ✅ | ❌ | ❌ | ❌ |
| **CRUD tipe kamar & unit fisik** | ✅ | ✅ | ❌ | ❌ |
| **Ubah harga & promo/voucher** | ✅ | ✅ | ❌ | ❌ |
| **Lihat semua booking** | ✅ | ✅ | ✅ | ❌ (hanya miliknya) |
| **Verifikasi bukti transfer** | ✅ | ✅ | ❌ | ❌ |
| **Check-in / Check-out** | ✅ | ✅ | ✅ | ❌ |
| **Ubah status kamar (dirty/maintenance)** | ✅ | ✅ | ✅ | ❌ |
| **Lihat laporan revenue & occupancy** | ✅ | ✅ | ❌ | ❌ |
| **Buat booking & upload bukti** | ❌ | ❌ | ❌ | ✅ |
| **Beri review** | ❌ | ❌ | ❌ | ✅ (hanya setelah check-out) |
| **Kelola profil sendiri** | ✅ | ✅ | ✅ | ✅ |

## Aturan Security (Anti IDOR/BOLA/BFLA)

| Ancaman | Contoh | Cara Cegah |
|---------|--------|------------|
| **IDOR** | Customer id=5 coba lihat booking id=99 milik orang lain via `/bookings/99` | Cek `booking.user_id == auth.id` ATAU role admin — di Policy |
| **BOLA** | Customer coba `POST /rooms` (buat kamar) | Middleware `role:manager,owner` — tolak 403 |
| **BFLA** | Receptionist coba buka `/reports/revenue` | Middleware `role:owner,manager` — tolak 403 |
| **Broken Auth** | Token dicuri | Password hash argon2/bcrypt, httpOnly cookie / JWT 15 menit + refresh, rate limit login |

**Contoh kode (Laravel Policy) — ilustrasi:**
```php
// BookingPolicy.php
public function view(User $user, Booking $booking): bool {
    return $user->id === $booking->user_id || $user->hasRole(['owner','manager','receptionist']);
}
public function verify(User $user): bool {
    return $user->hasRole(['owner','manager']);
}
```

## Edge Cases Role
| Kasus | Handling |
|-------|----------|
| Manager resign, akunnya masih aktif | Owner nonaktifkan user, semua booking tetap jalan (soft delete user, bukan hard delete) |
| Receptionist coba verifikasi booking (tidak berhak) | API return 403 + log `unauthorized attempt` |
| Customer coba check-in sendiri | Ditolak — hanya receptionist/manager/owner |
| Owner mau jadi customer juga | Boleh — 1 user 1 role utama, tapi bisa bikin akun customer terpisah (jangan campur role) |

## Best Practice
- **1 user = 1 role** di v1 (YAGNI — jangan many-to-many dulu).
- Simpan `role` di tabel `users.role` enum: `owner, manager, receptionist, customer`.
- Setiap endpoint cek **authentication dulu, baru authorization** (jangan kebalik).
- Log setiap aksi sensitif: `who, what, when, before, after` (audit log).

## Keputusan Grill — LOCKED 2026-10-08
1. **Setuju nama role di atas?** ✅ **Setuju** (Rafie 2026-10-08)
2. **Manager boleh lihat laporan uang?** ✅ **Ya, Manager + Owner**
3. **Customer perlu login dulu sebelum search?** ✅ **Tidak perlu login untuk search**, booking baru wajib login
