# 00 — Goal & Visi xyz-hotel

## Apa itu xyz-hotel?
Aplikasi booking hotel berbasis **website** dengan 2 sisi:

```
┌─────────────┐        ┌──────────────┐        ┌─────────────┐
│ Frontoffice │───────▶│  Backend API │◀───────│ Backoffice  │
│ (Pelanggan) │        │              │        │ (Manajemen) │
└─────────────┘        └──────────────┘        └─────────────┘
      │                                              │
  Lihat kamar,                           Kelola kamar, harga,
  cek tanggal,                          verifikasi booking,
  booking, upload                       laporan, voucher,
  bukti bayar                           ganti status kamar
```

**Analogi membumi:**
- Frontoffice = seperti kamu buka Traveloka → cari hotel → pilih tanggal → pesan
- Backoffice = seperti dashboard resepsionis di lobi hotel → lihat siapa check-in hari ini, kamar mana kotor/bersih

## Tujuan Utama
1. Pelanggan bisa booking kamar **tanpa telpon resepsionis**.
2. Manajemen bisa kelola kamar, harga, booking, dan laporan **tanpa Excel manual**.
3. Portfolio pribadi Rafie — bebas kreatif, kualitas tinggi, no deadline, jadi showcase skill.

## Batasan Fase 0 (yang sudah disepakati 2026-10-08)
| Aspek | Keputusan | Alasan |
|-------|-----------|--------|
| Jumlah hotel | **Satu hotel dulu** | YAGNI — jangan over-engineering multi-hotel di awal, tapi DB siap scale |
| Level fitur | **Lengkap** (search, availability, booking, review, voucher, laporan) | Portfolio harus wow, tapi bertahap implementasinya |
| Role | **4 role** | Butuh pemisahan tanggung jawab yang jelas |
| Pembayaran | **Manual transfer + upload bukti** | Simple, no biaya gateway, cocok untuk portfolio lokal |
| Availability | **WAJIB real-time** | Inti hotel — kalau salah, bisa double booking |
| Tech stack | **TIDAK pakai Next.js** | User big no, hormati preferensi |
| Environment | **Local only di MB Air M4** | Postgres 18.4 / MySQL 9.6, PHP 8.2, Node 26, tanpa Docker |
| Repo | Local `main` branch, nanti push ke GitHub | No worktrees, langsung di local branch |

## Yang BUKAN Tujuan (Non-Goal) di v1
- ❌ Payment gateway otomatis (Midtrans/Xendit) — v2
- ❌ Multi-hotel / multi-cabang — v2
- ❌ Mobile app native — fokus website dulu
- ❌ Channel manager (sinkron ke Agoda/Booking.com) — out of scope

## Success Criteria (Kapan dibilang berhasil?)
- Pelanggan bisa selesaikan booking dari cari tanggal sampai upload bukti tanpa bantuan admin.
- Admin bisa verifikasi booking < 5 menit dan ubah status kamar.
- Tidak pernah terjadi double booking di tanggal yang sama (ACID terjaga).
- Laporan harian/bulanan keluar akurat.

## Definition of Done Fase 0
- [x] Folder `xyz-hotel` + git init
- [x] 6 file docs ini terisi
- [ ] Tech stack final tanpa Next.js disepakati
- [ ] 4 role + alur booking disepakati
- [ ] Siap ke Fase 1 scaffolding

> Next: baca `01-brainstorming-log.md` untuk konteks percakapan lengkap.
