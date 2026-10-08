# 01 — Brainstorming Log (Sesi 2026-10-08)

> File ini merekam **keseluruhan percakapan & konteks** Fase 0. Tujuan: kalau besok lupa, buka file ini langsung ingat.

## Ringkasan Sesi

| Waktu | Aktor | Isi Penting |
|-------|-------|-------------|
| 2026-10-08 | Rafie | Mau bikin aplikasi booking hotel website: frontoffice (pelanggan) + backoffice (manajemen). Nama project `xyz-hotel` di `/Users/rafiesafarazaribowo/Projects`. Minta interaktif satu-satu, super presisi, bahasa Indonesia membumi, visual, contoh, tradeoff, edge cases, no shortcut/happy-path, perhatikan security, clean code, SOLID, dll. Tanya lebih detail dan mendalam. |
| 2026-10-08 | Agent | Audit faktual dulu: cek folder Projects ADA, xyz-hotel BELUM ADA (aman), posisi di /tmp, Node 26, Bun 1.4, PHP 8.2, Postgres 18.4, MySQL 9.6, Git 2.53, Docker TIDAK ADA. Pola referensi `xyz-haircut` = Dart Frog + Flutter monorepo. Tanya 6 pertanyaan: Q1 stack, Q2 level fitur, Q3 role, Q4 pembayaran, Q5 tipe kamar, Q6 availability. Kasih tradeoff tabel + DoD + skill mapping. |
| 2026-10-08 | Rafie | Jawab: Fase 0 banget, belum tau stack/flow, mau grill + dokumentasi ke 5+ file .md, pakai MB Air M4, local dulu, satu hotel, Q1 BIG NO Next.js, Q2 level lengkap, Q3 4 role, Q4 manual transfer, Q5 bebas kreatif (portfolio pribadi, no deadline/bos/client), Q6 perlu availability, nanti bikin repo GitHub. |
| 2026-10-08 | Agent | Eksekusi Fase 0: buat folder + git init (branch main), tulis AGENTS.md, README, .gitignore, dan 6 file docs. Grill Q1 tanpa Next.js. |

## Keputusan yang Sudah Lock (Jangan Diutak-atik Tanpa Diskusi)
1. **Satu hotel** — bukan multi-hotel di v1.
2. **Level lengkap** — tapi implementasi bertahap, tidak langsung semua.
3. **4 role** — detail di `04-roles-permissions.md`.
4. **Manual transfer** — upload bukti + verifikasi admin.
5. **Availability wajib** — harus ACID, cegah double booking.
6. **No Next.js** — hormati preferensi user.
7. **Local-first di M4** — tanpa Docker, pakai Postgres/MySQL lokal.
8. **Portfolio bebas** — no deadline, no atasan, kualitas > kecepatan.

## Pertanyaan yang Masih Menggantung (Perlu Grill Lanjutan)
- Q1: Stack pengganti Next.js mau apa? (Laravel vs Nuxt vs SvelteKit vs Dart Frog) — lihat `03-tech-stack-decision.md`
- Q3: 4 role itu siapa saja? Apa izin tiap role? — lihat `04-roles-permissions.md`
- Q5: Tipe kamar & harga — agent atur bebas, tapi perlu validasi user
- Q6: Aturan availability — berapa malam minimal, check-in/out jam berapa?

## Audit Faktual (Hasil `bash` 2026-10-08)
```
Projects EXISTS (7 project lama)
xyz-hotel NOT_EXISTS → aman dibuat
Node v26.7.0, npm 11.19.0, Bun 1.4.0
PHP 8.2.30
Postgres 18.4, MySQL 9.6, Docker NOT FOUND
Git 2.53.0
xyz-haircut = Dart Frog backend + Flutter app (referensi)
MB Air M4, local only
```

## Gaya Kerja yang Disepakati
- Interaktif iteratif satu-satu, jangan dump semua steps sekaligus.
- Selalu audit validasi crosscheck sebelum action.
- Instruksi super presisi, sedetail & sespesifik mungkin.
- Bahasa Indonesia membumi, untuk pemula, dengan contoh & visual.
- Wajib tradeoff, edge cases, best practice, opsi terbaik.
- Haram shortcut/workaround/happy-path-only.
- Perhatikan scalability, maintainability, security (IDOR, BOLA, SQLi, XSS, CSRF, dll).
- Clean Code, SRP, Error Handling + Logging, testable, dokumentasi API jelas.
- YAGNI, DRY, KISS, SOLID, ACID, no spaghetti.
- Manfaatkan yang sudah ada, jangan handwritten from scratch.
- No worktrees, local branch, no visual companion, make no mistakes.

## File yang Dibuat di Sesi Ini
- `AGENTS.md`
- `README.md`
- `.gitignore`
- `docs/00-goal.md`
- `docs/01-brainstorming-log.md` (file ini)
- `docs/02-prd.md`
- `docs/03-tech-stack-decision.md`
- `docs/04-roles-permissions.md`
- `docs/05-booking-flow-and-edge-cases.md`
- `docs/06-context-and-constraints.md`

> Semua file di atas adalah **sumber kebenaran (source of truth)** Fase 0. Kalau ada konflik, tanya Rafie dulu, jangan asumsi.
