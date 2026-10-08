# 05 — Booking Flow & Edge Cases (Jantung Hotel)

> Grill Q6 — Availability wajib. Ini bagian paling rawan, harus ACID.

## Alur Normal (Happy Flow) — Visual

```
Tanggal: 10-12 Okt (2 malam), Tipe: Deluxe, Unit total: 5

Customer                Frontend                Backend (DB Transaction)         Admin
   │                       │                           │                          │
   ├─ pilih tgl 10-12 ───▶│                           │                          │
   │                       ├─ GET /availability ──────▶│                          │
   │                       │   ┌─ hitung:              │                          │
   │                       │   │  SELECT COUNT(*)      │                          │
   │                       │   │  FROM bookings        │                          │
   │                       │   │  WHERE room_type=Deluxe│                          │
   │                       │   │  AND status IN (verified, checked_in)            │
   │                       │   │  AND overlap(tgl)     │                          │
   │                       │◀──┤  → 3 terisi, 2 kosong │                          │
   │◀─ tampil 2 kosong ────┤                           │                          │
   ├─ booking ────────────▶├─ POST /bookings ─────────▶│                          │
   │                       │   ┌─ BEGIN TRANSACTION     │                          │
   │                       │   │  1. cek availability lagi (FOR UPDATE)           │
   │                       │   │  2. insert booking pending_payment               │
   │                       │   │  3. COMMIT / ROLLBACK  │                          │
   │                       │◀──┤  → pending_payment     │                          │
   │◀─ instruksi transfer ─┤                           │                          │
   ├─ upload bukti ───────▶├─ POST /bookings/1/proof ─▶│                          │
   │                       │                           ├─ waiting_verification ──▶│
   │                       │                           │◀─ approve ───────────────┤
   │◀─ verified + invoice ─┤                           │                          │
   │                       │                           │                          ├─ check-in → occupied
   │                       │                           │                          ├─ check-out → dirty → available
```

**Status Booking:**
```
pending_payment → waiting_verification → verified → checked_in → checked_out
       │                    │               │
       ├─ expired (2 jam)   └─ rejected ────┘
       └─ cancelled (oleh customer sebelum verifikasi)
```

## Aturan Overlap (Kunci Cegah Double Booking)

**Definisi overlap:** Booking A `[10,12)` dan Booking B `[11,13)` → overlap di tgl 11.

**Rumus DB (Postgres/MySQL):**
```sql
-- Cek apakah ada overlap dengan booking yang sudah verified/checked_in
WHERE room_type_id = :type
  AND status IN ('verified','checked_in')
  AND check_in < :new_check_out
  AND check_out > :new_check_in
```

**Contoh hitung:**
- Deluxe total unit = 5
- Sudah ada 5 booking verified yang overlap tgl 11 → **tolak** booking baru (403, "Kamar penuh")
- Hanya 4 yang overlap → **boleh** (1 slot tersisa)

## Edge Cases Ekstrem (Wajib Di-handle, Haram Happy-Path-Only)

| # | Edge Case | Contoh | Handling |
|---|-----------|--------|----------|
| 1 | **Race condition** — 2 orang booking detik yang sama | Jam 10:00:00, Ani & Budi booking Deluxe tgl sama, sisa 1 slot | Pakai `DB::transaction()` + `SELECT ... FOR UPDATE` / unique constraint. Satu berhasil, satu gagal dengan pesan jelas. |
| 2 | **Booking expired** | Ani booking tapi tidak upload bukti 2 jam | Cron/job tiap menit: `pending_payment` & `created_at < now-2h` → `expired`. Slot kembali kosong. |
| 3 | **Upload bukti palsu / salah** | Upload foto kosong, atau nominal kurang | Admin reject + alasan, status `rejected`, customer bisa upload ulang (max 3x) atau booking `cancelled`. |
| 4 | **Check-in tanpa verifikasi** | Resepsionis iseng check-in booking pending | Ditolak — hanya `verified` bisa `checked_in`. Validasi di backend, jangan di frontend saja. |
| 5 | **Check-out terlambat** | Tamu check-out jam 14:00 (lewat 12:00) | Status tetap `checked_in`, denda hitung manual di backoffice (catat di laporan). |
| 6 | **Batal setelah verified** | Tamu verified tapi mau batal H-1 | Manager cancel → status `cancelled`, unit kembali kosong, refund manual (catat). |
| 7 | **Voucher expired / quota habis** | Kode PROMO10 dipakai 101x padahal quota 100 | Cek voucher `expiry` & `used_count < quota` dalam transaction yang sama. |
| 8 | **Tanggal invalid** | check_in = check_out, atau check_in di masa lalu | Validasi: `check_out > check_in`, `check_in >= today`, `malam >=1`. |
| 9 | **Kapasitas berlebih** | Deluxe kapasitas 2 orang, tamu isi 5 orang | Validasi `guests <= capacity` + `guests >=1`. |
| 10 | **Upload file berbahaya** | Upload `.php` atau 50MB | Validasi MIME `jpg,png,pdf`, max 5MB, rename file, simpan di `storage/app/private` (bukan public), scan. |
| 11 | **IDOR lihat booking orang lain** | Customer A coba GET /bookings/99 milik B | Policy: `booking.user_id == auth.id` atau role admin. |
| 12 | **Harga diubah saat booking berjalan** | Harga Deluxe naik dari 500k ke 600k setelah Ani booking | Harga di-**snapshot** ke `bookings.total_price` saat create — tidak ikut harga baru. |

## Best Practice ACID & Logging

```php
// Ilustrasi Laravel — WAJIB transaction
DB::transaction(function () use ($data) {
    // 1. Lock baris yang relevan
    $occupied = Booking::where('room_type_id', $data->type)
        ->whereIn('status', ['verified','checked_in'])
        ->where('check_in', '<', $data->check_out)
        ->where('check_out', '>', $data->check_in)
        ->lockForUpdate()->count();

    $totalUnits = RoomType::find($data->type)->total_units;
    if ($occupied >= $totalUnits) {
        throw new \Exception('Kamar penuh di tanggal tersebut');
    }

    // 2. Insert booking
    $booking = Booking::create([... , 'total_price' => $priceSnapshot]);

    // 3. Log audit
    AuditLog::create(['user_id' => auth()->id(), 'action' => 'booking_created']);

    return $booking;
});
```

- **Logging:** Setiap perubahan status log: `booking_id, from, to, by, at, reason`.
- **Error handling:** Jangan `try-catch` kosong — log ke `storage/logs/laravel.log` + return pesan user-friendly.

## Pertanyaan Grill untuk Kamu
1. **Expired berapa jam?** Rekomendasi: **2 jam** untuk `pending_payment` (contoh Traveloka). Setuju?
2. **Check-in/out jam berapa?** Rekomendasi: **14:00 / 12:00**. Setuju?
3. **1 booking boleh berapa kamar?** Rekomendasi: **1 tipe, 1 unit, N malam** (YAGNI — jangan multi-kamar dulu). Setuju?

> Jawab 3 poin itu, lalu kita lock flow-nya.

> Next: `06-context-and-constraints.md` untuk konteks portfolio & local-first.
